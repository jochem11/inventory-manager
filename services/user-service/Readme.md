# user-service

Owns the **user profiles**: name, email, phone and avatar. Other services never read its database. The gateway calls it over gRPC, and it exchanges events with the other services over Kafka.

It doesn't know about passwords or logins; those belong to the [auth-service](../auth-service/Readme.md). A profile and its login share the same id: the auth-service creates the id at registration, and this service creates the profile under that id when it receives the `IdentityRegistered` event.

![Building blocks](docs/overview.svg)

The diagrams are pages of [docs/user-service.drawio](docs/user-service.drawio). See [Editing the diagrams](#editing-the-diagrams).

## Entry points

There are three ways into the service, all started in [cmd/main.go](cmd/main.go):

| Entry point | What comes in | Code |
|---|---|---|
| gRPC server, `:50051` | `UserService` calls from the gateway | [internal/grpc](internal/grpc) |
| Kafka consumer, group `user-service` | `auth.events`: `IdentityRegistered` | [internal/events](internal/events) |
| Outbox relay (background loop) | unpublished rows in the `outbox` table, published to `user.events` | [shared/kafka](../../shared/kafka/outbox.go) |

At startup, `main.go` does the following, in order:
1. Sets up tracing to Jaeger.
2. Connects to MySQL and migrates the tables.
3. Starts the outbox relay and the Kafka consumer.
4. Serves gRPC, including health checks and reflection, so `grpcurl` works without the `.proto` files.

### gRPC API ([proto/user/user.proto](../../proto/user/user.proto))

| RPC | Does |
|---|---|
| `GetUser(id)` | One user, `NOT_FOUND` when missing or deleted |
| `GetUsers(ids)` | Several users by id in one call; unknown ids are skipped |
| `ListUsers(offset, limit, order_by, sort_direction, filter)` | One page of users plus the total; see [ListUsers](#listusers) |
| `CreateUser(input)` | A profile without a login (admin use) |
| `UpdateUser(id, input)` | Replaces every editable field |
| `DeleteUser(id)` | Soft delete, and publishes `UserDeleted` |

Errors come back as gRPC status codes:
- `INVALID_ARGUMENT`: invalid input, with a `BadRequest` detail per field, so a client can show each message next to its field.
- `NOT_FOUND`: no such user.
- `ALREADY_EXISTS`: the email is taken.
- `INTERNAL`: anything else. The cause is logged, not sent to the client.

### Events ([proto/user/events.proto](../../proto/user/events.proto))

| Direction | Topic | Event |
|---|---|---|
| consumes | `auth.events` | `IdentityRegistered`: creates the profile |
| publishes | `user.events` | `UserDeleted`: the auth-service removes the login |

The record key is the user id, so all events about one user stay in order.

## How a call flows through the code

```
gRPC handler (internal/grpc)       proto → types, domain error → status code
  → UserService (service/)         validation (internal/validation), business rules
    → UserRepository (internal/repository)  GORM queries, paging, the outbox
```

The packages depend on interfaces in [internal/domain](internal/domain), not on each other's code. That makes them easy to replace in tests.

**Validation** uses struct tags. `mod` tags clean the input up (trim, lower-case the email). `validate` tags then check it (required, max length, E.164 phone, http(s) URL). See [internal/models/user.go](internal/models/user.go) and [pkg/types/types.go](pkg/types/types.go).

## Workflows

### ListUsers

![ListUsers](docs/list-users.svg)

The users table in the web app calls this through the gateway. The steps:

1. **Validate.** The service checks the paging values (`offset` ≥ 0, `limit` 1–100, default 10) and that the sort field is known.
2. **Order.** [shared/paging](../../shared/paging/paging.go) turns the sort field into a column, using an allow-list so no user input reaches the SQL. It always adds `id` as the last sort key, so rows with equal values keep a stable order across pages.
3. **Query.** One query returns both the page and the total: `COUNT(*) OVER()` counts all matches before `LIMIT` is applied.
4. **Fallback.** Only for a page past the end, which has no rows to read the total from, a separate `COUNT` runs.

Filters are `LIKE` patterns, with `%` and `_` in the input escaped. They ignore case because the database uses a case-insensitive collation.

### UpdateUser

![UpdateUser](docs/update-user.svg)

Update replaces the whole profile. A field you leave out, like `phone`, is cleared. The new values are validated just like on create. If the email changes, the service checks that no other user has it. The unique index on `email` is the final guard, for two requests at the same moment.

### Creating the profile when someone registers

![IdentityRegistered](docs/profile-from-event.svg)

Kafka delivers each event **at least once**, so the same event can arrive twice. The consumer handles each event inside one transaction:
1. It inserts the event id into `processed_events`.
2. It creates the profile.

If the event id is already there, the event was handled before and is skipped. Because both writes commit together, a redelivered event can't create a second profile.

What happens when an event fails depends on why:
- **Data no retry can fix** (invalid fields, email already taken): the event is logged and skipped. This is `kafka.Permanent`.
- **Anything else** (for example MySQL is down): the event is retried, waiting longer each time, from 1 s up to 30 s. The Kafka offset is only committed after the event has been handled.

### DeleteUser

![DeleteUser](docs/delete-user.svg)

The **transactional outbox**: the soft delete and the `UserDeleted` event are written in one database transaction, so there can't be a deleted user without an event, or an event without a deletion. The relay publishes the event to Kafka within about half a second. If Kafka is down, events wait in the table until it's back.

Published outbox rows are kept for 7 days, which helps when debugging, and then deleted.

## Data model

![Data model](docs/data-model.svg)

IDs are [KSUIDs](https://github.com/segmentio/ksuid): 27 characters that sort by creation time. A deleted user is only marked with `deleted_at`. Its email stays in the unique index, so it can't be used to register again (a known limitation).

## Configuration

| Env var | Default | |
|---|---|---|
| `GRPC_ADDR` | `:50051` | |
| `DB_HOST`, `DB_PORT` | `localhost`, `3306` | In Tilt: the MySQL on your machine |
| `DB_USER`, `DB_PASSWORD` | `root`, none | Set in `.env` (see [.env.example](../../.env.example)) |
| `DB_NAME` | `user_service` | |
| `KAFKA_BROKERS` | `127.0.0.1:9094` | Comma-separated; in Tilt `kafka:9092` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | unset: no tracing | In Tilt `http://jaeger:4317` |

## Running

Tilt runs everything; see the [Tiltfile](../../Tiltfile). To run only this service on your machine, with Kafka and MySQL already running:

```sh
make run     # gRPC on :50051
make test    # repository tests use MySQL (DB_* env) and skip without it
make proto   # after changing proto/user/*.proto or proto/auth/events.proto
make         # list all targets
```

Try it with [grpcurl](https://github.com/fullstorydev/grpcurl):

```sh
grpcurl -plaintext -d '{"limit": 5}' localhost:50051 user.UserService/ListUsers
```

## Layout

```
cmd/main.go            wiring: tracing, database, Kafka, gRPC server
service/               UserService: the business rules
internal/
  domain/              interfaces + errors the other packages share
  grpc/                gRPC handler, proto ⇄ types mapping, error → status
  repository/          GORM queries; Delete also writes the outbox
  events/              Kafka consumer for auth.events
  models/              GORM models with mod/validate tags
  validation/          runs the mod + validate tags
pkg/types/             request types (list params, filters, input)
pkg/pb/                protoc output, never edit
docs/                  user-service.drawio + one SVG per page
```

## Editing the diagrams

Open [docs/user-service.drawio](docs/user-service.drawio) in [draw.io](https://app.diagrams.net) or the VS Code extension *Draw.io Integration*. After a change, export each page as SVG over the matching file in `docs/`: *File → Export as → SVG*, with *Include a copy of my diagram* on. Each SVG also contains its own page, so draw.io can open an SVG directly too.
