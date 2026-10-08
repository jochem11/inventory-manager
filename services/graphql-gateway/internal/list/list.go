// Package list holds what every list query shares:
// things(offset, limit, orderBy, filter).
package list

import (
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/gqlerr"
	"github.com/jochem11/inventory-manager/services/graphql-gateway/internal/graph/model"
)

const maxLimit = 100

// ValidatePage checks offset ≥ 0 and 1 ≤ limit ≤ 100.
func ValidatePage(offset, limit int) error {
	if offset < 0 {
		return gqlerr.BadUserInput("offset", "offset must be 0 or more")
	}
	if limit < 1 || limit > maxLimit {
		return gqlerr.BadUserInput("limit", "limit must be between 1 and 100")
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
