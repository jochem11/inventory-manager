// Package list holds what every list query shares:
// things(offset, limit, orderBy, filter).
package list

import (
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/graph/model"
	"github.com/jochem11/inventory-manager/shared/errs"
)

const maxLimit = 100

// ValidatePage checks offset ≥ 0 and 1 ≤ limit ≤ 100.
func ValidatePage(offset, limit int) error {
	if offset < 0 {
		return errs.Field("offset", "offset must be 0 or more")
	}
	if limit < 1 || limit > maxLimit {
		return errs.Field("limit", "limit must be between 1 and 100")
	}
	return nil
}

// SortDirection converts the GraphQL sort direction to the "asc"/"desc" the
// services' List requests take.
func SortDirection(d model.SortDirection) string {
	if d == model.SortDirectionDesc {
		return "desc"
	}
	return "asc"
}
