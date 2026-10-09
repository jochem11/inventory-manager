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
// Columns are qualified with the query's own table, so the order also works
// on a query that joins other tables.
func OrderBy[F comparable](columns map[F]string, field F, direction SortDirection) (clause.OrderBy, error) {
	var order clause.OrderBy
	var none F
	if field != none {
		column, ok := columns[field]
		if !ok {
			return clause.OrderBy{}, fmt.Errorf("unsupported sort field %v", field)
		}
		order.Columns = append(order.Columns, clause.OrderByColumn{
			Column: clause.Column{Table: clause.CurrentTable, Name: column},
			Desc:   direction == SortDesc,
		})
	}
	order.Columns = append(order.Columns, clause.OrderByColumn{Column: clause.Column{Table: clause.CurrentTable, Name: "id"}})
	return order, nil
}

// Params selects a page: embed it in a list's parameters, next to its
// ordering and filter. Clean (shared/validation) defaults the limit to 10.
type Params struct {
	Offset int `json:"offset" validate:"min=0"`
	Limit  int `json:"limit" mod:"default=10" validate:"min=1,max=100"`
}

// Find loads one page of query's rows and the total number of matches. query
// must have its Model and filters set, but no Order, Offset or Limit. It may
// join relations (query.Joins("Category")); they're filled in on each row.
// Filters on a joining query must name their table: "items.name LIKE ?".
//
// COUNT(*) OVER() runs after WHERE but before LIMIT, so every row carries the
// total: the page and the total come from one query. Only a page past the end
// has no rows to read it from, and then a separate COUNT runs.
func Find[T any](query *gorm.DB, order clause.OrderBy, params Params) (*Page[*T], error) {
	// A new session, so the filters can be reused by the COUNT below.
	query = query.Session(&gorm.Session{})

	var rows []struct {
		Row        T `gorm:"embedded"`
		TotalCount int64
	}
	// The table's own columns (with joins, a bare * would repeat id, name and
	// so on), the joined relations' columns, and the total.
	joined, err := joinedColumns(query)
	if err != nil {
		return nil, err
	}
	columns := strings.Join(append([]string{"?.*"}, joined...), ", ")
	if err := query.Select(columns+", COUNT(*) OVER() AS total_count", clause.Table{Name: clause.CurrentTable}).
		Order(order).Offset(params.Offset).Limit(params.Limit).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	page := &Page[*T]{Nodes: make([]*T, len(rows))}
	for i := range rows {
		page.Nodes[i] = &rows[i].Row
	}
	if len(rows) > 0 {
		page.TotalCount = rows[0].TotalCount
	} else if params.Offset > 0 {
		if err := query.Count(&page.TotalCount).Error; err != nil {
			return nil, err
		}
	}
	return page, nil
}

// joinedColumns selects the columns of the relations query joins with
// Joins("Relation"), aliased the way GORM fills them in: Relation__column.
// GORM adds these itself, except when the query has its own Select, as Find
// does.
func joinedColumns(query *gorm.DB) ([]string, error) {
	stmt := query.Statement
	if len(stmt.Joins) == 0 {
		return nil, nil
	}
	if err := stmt.Parse(stmt.Model); err != nil {
		return nil, err
	}
	var columns []string
	for _, join := range stmt.Joins {
		relation, ok := stmt.Schema.Relationships.Relations[join.Name]
		if !ok {
			continue // a raw SQL join, not a relation
		}
		for _, column := range relation.FieldSchema.DBNames {
			columns = append(columns, fmt.Sprintf("%s.%s AS %s",
				stmt.Quote(join.Name), stmt.Quote(column), stmt.Quote(join.Name+"__"+column)))
		}
	}
	return columns, nil
}

// Contains turns input into a LIKE pattern matching it anywhere, escaping the
// wildcards so "%" and "_" are matched literally. With MySQL's default
// case-insensitive collation, LIKE ignores case.
func Contains(s string) string {
	s = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
	return "%" + s + "%"
}
