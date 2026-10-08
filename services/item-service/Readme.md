# item-service

> **Status: in progress.** The data model, the domain interfaces and the item-status repository are built. The gRPC API, the services and the deployment are next. Until then, the web app's items page runs on mock data.

Will own the **inventory**:
- **items**, with their categories and statuses
- **locations:** where things are, from building down to a single spot
- **assignments:** which items each person has, with the full history

| | |
|---|---|
| **API** | gRPC on `:50053` (planned), `proto/item/item.proto` |
| **Database** | MySQL `item_service` |
| **Consumes** | `user.events`: `UserDeleted` (planned) |

![Overview](docs/overview.svg)

## Roadmap

![Roadmap sketch](docs/roadmap.svg)

The plan for tying items to people and places. Solid boxes exist today; dashed ones are planned and may still change.

- **Locations form a tree:** building → floor → room → spot. A location can carry a GeoJSON shape, which feeds the floor plan in the web app.
- **An item has one current location and one current holder.** Every change adds an `Assignment` row, so the history of who had what, and where, is kept.
- **Users stay in the user-service.** The item-service stores only their id; the gateway fills in names when a query asks for them. When a user is deleted (`UserDeleted`), their open assignments are closed.

### Assigning an item (planned)

![Assign an item](docs/assign-item.svg)

1. The gateway checks the permission (`items:write`) and that the user exists.
2. The item-service closes the current assignment and opens the new one, in one transaction.

### Milestones

| | Milestone | Status |
|---|---|---|
| 1 | Data model: items, categories, statuses | Done |
| 2 | Repositories with paging, sorting and filtering | Item statuses done; items and categories next |
| 3 | Services, validation, gRPC API, `cmd/main.go` | Planned |
| 4 | Deployment in Tilt, gateway schema, web app on the real API | Planned |
| 5 | Locations, synced with the floor plan | Planned |
| 6 | Assignments and history; `UserDeleted` handling | Planned |

## Data model today

![Data model](docs/data-model.svg)

- **Every item has a category and a status.** Both use `ON DELETE RESTRICT`, so a category or status that items still use can't be deleted.
- **Names are unique and ignore case:** *Tools* and *tools* count as the same category.
- **Validation is built into the models** ([internal/models](internal/models)): `mod` tags clean input up, `validate` tags check it, and `json` tags name the fields in error messages.

## Listing

![List item statuses](docs/list-item-statuses.svg)

Listing uses [shared/paging](../../shared/paging/paging.go), like the user-service:
- `OrderBy` maps sort fields to columns through an allow-list.
- `Find` loads a page and its total in one query.
- `Contains` builds case-insensitive substring filters.

Items, categories and locations will follow [ItemStatusRepository](internal/repository/item_status_repository.go).

## Project structure

```
internal/
  domain/              service and repository interfaces, errors
  models/              Item, Category, ItemStatus (GORM + validation tags)
  repository/          ItemStatusRepository
  grpc/, events/, validation/   planned
pkg/types/             list params, filters, sort fields, input
docs/                  item-service.drawio and its SVGs
```
