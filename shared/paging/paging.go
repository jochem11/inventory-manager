// Package paging holds what every list query shares: the page type, sorting
// with a stable tie-breaker, LIKE patterns, and loading a page together with
// its total in one query.
package paging

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SortDirection string

const (
	SortAsc  SortDirection = "ASC"
	SortDesc SortDirection = "DESC"
)

// Page is one page of a list, plus how many items match across all pages.
type Page[T any] struct {
	TotalCount int64 `json:"totalCount"`
	Nodes      []T   `json:"nodes"`
}

// OrderBy builds the ORDER BY for a list query. columns maps the sort fields a
// list allows to their columns; only these reach SQL, so a sort field can't
// inject anything. An empty field means no sort was asked for.
//
// The ID is always the last key: a unique tie-breaker keeps pages stable when
// the sorted column has duplicates, and KSUIDs roughly follow creation time.
func OrderBy[F comparable](columns map[F]string, field F, direction SortDirection) (clause.OrderBy, error) {
	var order clause.OrderBy
	var none F
	if field != none {
		column, ok := columns[field]
		if !ok {
			return clause.OrderBy{}, fmt.Errorf("unsupported sort field %v", field)
		}
		order.Columns = append(order.Columns, clause.OrderByColumn{
			Column: clause.Column{Name: column},
			Desc:   direction == SortDesc,
		})
	}
	order.Columns = append(order.Columns, clause.OrderByColumn{Column: clause.Column{Name: "id"}})
	return order, nil
}

// Find loads one page of query's rows and the total number of matches. query
// must have its Model and filters set, but no Order, Offset or Limit.
//
// COUNT(*) OVER() runs after WHERE but before LIMIT, so every row carries the
// total: the page and the total come from one query. Only a page past the end
// has no rows to read it from, and then a separate COUNT runs.
func Find[T any](query *gorm.DB, order clause.OrderBy, offset, limit int) (*Page[*T], error) {
	// A new session, so the filters can be reused by the COUNT below.
	query = query.Session(&gorm.Session{})

	var rows []struct {
		Row        T `gorm:"embedded"`
		TotalCount int64
	}
	if err := query.Select("*, COUNT(*) OVER() AS total_count").
		Order(order).Offset(offset).Limit(limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	page := &Page[*T]{Nodes: make([]*T, len(rows))}
	for i := range rows {
		page.Nodes[i] = &rows[i].Row
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	} else if offset > 0 {
		if err := query.Count(&page.TotalCount).Error; err != nil {
			return nil, err
		}
	}
	return page, nil
}

// Contains turns input into a LIKE pattern matching it anywhere, escaping the
// wildcards so "%" and "_" are matched literally. With MySQL's default
// case-insensitive collation, LIKE ignores case.
func Contains(s string) string {
	s = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
	return "%" + s + "%"
}
