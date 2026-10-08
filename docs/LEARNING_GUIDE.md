# Build it yourself: GraphQL gateway → gRPC services → Kafka + Postgres

A hands-on guide for building the backend in the diagrams in this folder, one milestone at a time. It tells you **what** to build, **why**, and how to check that it works, but it isn't a copy-paste solution. Most snippets are hints, not finished files.

Keep these diagrams open while you work:

| Diagram | Use it for |
|---|---|
| [architecture.drawio](architecture.drawio) | The big picture. Come back to it whenever you're lost. |
| [graphql-api.drawio](graphql-api.drawio) | Milestones 4 and 5 (paging, sorting, filtering) |
| [request-flow.drawio](request-flow.drawio) | Milestones 4 and 6 |
| [data-model.drawio](data-model.drawio) | Milestones 5 and 7 |
| [event-flow.drawio](event-flow.drawio) | Milestone 7 |
| [auth-workflows.drawio](auth-workflows.drawio) | Login, JWTs and the Kafka outbox as built: UML sequence, activity, state machine and deployment diagrams |

## How to use this guide

- **Go in order.** Each milestone ends in a working system. Get it working first and improve it later.
- **Start in memory, then add the hard parts.** You'll build gRPC first, then GraphQL, then Postgres, then Kafka, then Kubernetes. Learning five things at once makes debugging impossible.
- **Run things on your machine first** (`cargo run`, `go run`), and only move them into Tilt/k8s in milestone 8. The exception is Postgres and Kafka, which you'll run in the cluster and reach through port-forwards.
- **Use the checkpoints.** Each milestone has a "Checkpoint" section. Don't move on until it passes.
- Library APIs change. When a hint doesn't compile, check the library's current docs and examples; working that out is part of the learning.

## Milestone overview

| # | You build | You learn |
|---|---|---|
| 0 | Tooling | The toolchain |
| 1 | `.proto` contracts | Protobuf, API design, code generation |
| 2 | user-service (Go, in memory) | gRPC servers in Go, reflection, grpcurl |
| 3 | item-service (Rust, in memory) | tonic, prost, async Rust |
| 4 | graphql-gateway (Go) | GraphQL schemas, gqlgen, resolvers, gRPC clients, error mapping |
| 5 | Postgres per service | Migrations, dynamic SQL safely, offset paging, indexes |
| 6 | Wire up the web app | Swapping a mock for a real API without changing the UI |
| 7 | Kafka and the outbox | Event-driven design, at-least-once delivery, idempotency, projections |
| 8 | Containers, k8s and Tilt | Multi-stage Dockerfiles, Deployments, Services, dependency ordering |
| 9 | Stretch goals | DataLoader, auth propagation, tracing, tests, contract checks |

---

## Milestone 0: Tooling

Install these tools and check that each one runs:

| Tool | Why | Check |
|---|---|---|
| Rust (`rustup`) | item-service | `cargo --version` |
| Go | user-service and gateway | `go version` |
| `protoc` | code generation from `.proto` files (Go and Rust) | `protoc --version` |
| [buf](https://buf.build/docs) | proto linting, breaking-change checks, Go codegen | `buf --version` |
| `protoc-gen-go`, `protoc-gen-go-grpc` | Go plugins (`go install …@latest`) | on your `PATH` |
| [grpcurl](https://github.com/fullstorydev/grpcurl) | "curl for gRPC", for testing services by hand | `grpcurl --version` |
| `psql` | inspecting the databases | `psql --version` |
| `kcat` (optional) | reading Kafka topics from the terminal | `kcat -V` |
| Tilt, kubectl, a local cluster | you already have these for `web` | `tilt version` |

Rust's Kafka client (`rdkafka`) compiles a C library, so you'll also need `cmake` (`brew install cmake`).

---

## Milestone 1: Protobuf contracts

**Goal:** every service API and every event is defined once in `proto/`. Each language generates its own code from those files, and that's what lets the services be written in different languages.

**Learn:**
- proto3 syntax: messages, enums, `optional`, `repeated`, `oneof`, `google.protobuf.Timestamp`, `FieldMask`
- Package versioning (`item.v1`) and why you **never reuse or renumber field numbers**
- Enum conventions: the zero value is `*_UNSPECIFIED`, and names are prefixed (`ITEM_SORT_FIELD_NAME`)

**Build:**
```
proto/
  buf.yaml
  buf.gen.yaml
  common/v1/paging.proto   # SortDirection, IntFilter (shared)
  item/v1/item.proto       # ItemService
  item/v1/events.proto     # ItemEvent envelope
  user/v1/user.proto       # UserService
  user/v1/events.proto     # UserEvent envelope
```

1. `item.proto`: an `Item` message with the same fields as `Item` in [web/src/api/items.ts](../web/src/api/items.ts), plus `created_by`, `created_at` and `updated_at`.
2. `ItemService` with `GetItem`, `ListItems`, `CreateItem`, `UpdateItem`, `DeleteItems` and `DeleteAllItems`.
3. Model `ListItemsRequest` on the diagram in [graphql-api.drawio](graphql-api.drawio) (panel 3): `offset`, `limit`, `ItemOrder order_by`, `ItemFilter filter`. Make each filter field `optional`, so that "not set" is different from "empty string" or `0`.
4. `ListItemsResponse { repeated Item items; int32 total_count; }`.
5. `UserService` with `GetUser`, `GetUsers(repeated string ids)`, `ListUsers`, `CreateUser` and `DeleteUser`.
6. Events: one envelope per topic.
   ```proto
   message UserEvent {
     string event_id = 1;                       // uuid, used for deduplication
     google.protobuf.Timestamp occurred_at = 2;
     oneof payload {
       UserCreated created = 10;
       UserUpdated updated = 11;
       UserDeleted deleted = 12;
     }
   }
   ```

**Think about it:** why is `UpdateItem` given a `google.protobuf.FieldMask` (or all-`optional` fields)? What goes wrong if an update without a field mask sets `notes` to `""`?

**Checkpoint:** `buf lint` passes, and `buf generate` produces Go files for user-service.

---

## Milestone 2: user-service in Go (in memory)

**Goal:** your first gRPC server. It stores users in a `map` protected by a mutex, and grpcurl can call it.

**Learn:** [grpc-go quickstart](https://grpc.io/docs/languages/go/quickstart/), gRPC status codes, server reflection, the health checking protocol.

**Build:**
```
services/user-service/
  go.mod
  cmd/server/main.go        # listen on :50051, register services
  internal/server/server.go # implements userv1.UserServiceServer
  gen/                      # buf output
```

Hints:
- Embed `userv1.UnimplementedUserServiceServer` in your struct.
- Register `reflection.Register(s)` and the health server (`google.golang.org/grpc/health`). Reflection is what lets grpcurl work without `.proto` files.
- Return errors with `status.Error(codes.NotFound, "user not found")`, not plain `errors.New`. The gateway will rely on these codes later.
- `GetUsers` takes many ids and returns the users it finds, **silently skipping unknown ids**. The DataLoader in milestone 4 depends on this.

**Checkpoint:**
```sh
go run ./cmd/server
grpcurl -plaintext localhost:50051 list
grpcurl -plaintext -d '{"name":"Ada","email":"ada@example.com"}' localhost:50051 user.v1.UserService/CreateUser
grpcurl -plaintext -d '{"id":"does-not-exist"}' localhost:50051 user.v1.UserService/GetUser   # → NotFound
```

---

## Milestone 3: item-service in Rust (in memory)

**Goal:** the same idea in a second language. Seeing that grpcurl can't tell the two apart is the point of this milestone.

**Learn:** [tonic examples](https://github.com/hyperium/tonic/tree/master/examples), `build.rs`, `tokio`, `Arc<RwLock<…>>`.

**Build:**
```
services/item-service/
  Cargo.toml
  build.rs          # compiles ../../proto into Rust at build time
  src/main.rs
  src/service.rs    # impl ItemService for ItemServiceImpl
```

Hints:
- In recent tonic versions, code generation lives in the `tonic-prost-build` crate. Older versions used `tonic-build`. Match whatever the tonic README says for your version.
- In `build.rs`, generate only the server here (`build_client(false)`). The Go gateway generates its own clients.
- Include `common/v1/paging.proto` as well, and point the include path at the `proto/` root.
- Add `tonic-reflection` and `tonic-health` to get the same grpcurl experience as the Go service.
- Implement `ListItems` in memory for now: filter, sort, then `skip(offset).take(limit)`. Port the logic from the mock `fetchItems` in [web/src/api/items.ts](../web/src/api/items.ts). This gives you a reference behaviour to compare against when you switch to SQL.
- Seed the store from [web/src/data/items.json](../web/src/data/items.json) (`include_str!` + `serde_json`).

**Checkpoint:** run it on `:50052` so both services can run at the same time.
```sh
grpcurl -plaintext -d '{"offset":0,"limit":5,"orderBy":{"field":"ITEM_SORT_FIELD_QUANTITY","direction":"SORT_DIRECTION_DESC"}}' \
  localhost:50052 item.v1.ItemService/ListItems
```

---

## Milestone 4: The GraphQL gateway (Go)

**Goal:** one `/graphql` endpoint that translates to gRPC. The gateway has **no business logic**: it validates input, maps types and forwards calls.

**Learn:** [gqlgen](https://gqlgen.com/) (schema-first codegen), grpc-go clients, client interceptors, GraphQL error extensions.

**Build:**
```
services/graphql-gateway/
  go.mod
  Makefile                       # proto + gqlgen targets, like user-service's
  gqlgen.yml
  cmd/main.go                    # wiring: clients, routes, graceful shutdown
  schema/*.graphqls              # the SDL, the source of truth for the API
  pkg/pb/                        # protoc output for item/ and user/ (you only use the clients)
  internal/clients/              # grpc.NewClient + a call-logging interceptor
  internal/gqlerr/               # gRPC status → *gqlerror.Error, shared
  internal/list/                 # paging validation, shared
  internal/item/                 # item Resolver (calls item-service) + mapper.go
  internal/user/                 # user Resolver (calls user-service) + mapper.go
  internal/graph/generated/      # gqlgen output, never edit
  internal/graph/model/          # gqlgen output, plus your own models
  internal/graph/resolver.go     # Resolver struct holding one resolver per service
  internal/graph/*.resolvers.go  # gqlgen's resolvers, one-liners into internal/<service>
  internal/user/loader.go        # (milestone 9) UserLoader
```

Steps:
1. **Schema first.** Write the SDL from the "GraphQL API design" section of the plan, or read it off [graphql-api.drawio](graphql-api.drawio), into `schema/*.graphqls`. Run `go run github.com/99designs/gqlgen generate`. gqlgen writes the Go types for `Item`, `ItemFilter`, `ItemOrder` and `ItemSortField`, plus a stub for every resolver. Re-run it whenever the schema changes; your resolver bodies are kept.
2. **Proto code.** Copy the `proto` target from user-service's [Makefile](../services/user-service/Makefile) and point `PROTOS` at both `item/…` and `user/…`. protoc-gen-go-grpc generates clients and servers together; the gateway just never registers a server.
3. **Clients.** Create them once at startup from env vars (`ITEM_SERVICE_ADDR`, `USER_SERVICE_ADDR`) using `grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))`. This doesn't dial yet, so the gateway starts even when a service is down. Wrap each client in its service's resolver (`internal/item`, `internal/user`) and hold those in the `Resolver` struct in `internal/graph/resolver.go`. They're safe for concurrent use, so share them across requests.
4. **Resolvers.** For `items(offset, limit, orderBy, filter)`:
   - validate: `1 ≤ limit ≤ 100` and `offset ≥ 0`, otherwise return an error with code `BAD_USER_INPUT`
   - convert the GraphQL input into `ListItemsRequest` with small functions in `internal/item/mapper.go`, which keeps the resolvers tiny. Optional GraphQL inputs arrive as pointers (`*string`), and `optional` proto fields are pointers too, so "not set" survives the trip.
   - call gRPC with the resolver's `ctx` (so a cancelled HTTP request cancels the gRPC call) and map the response into `ItemConnection { totalCount, nodes }`
5. **Errors.** Write one function, used everywhere, that maps:
   ```go
   // hint
   switch status.Code(err) {
   case codes.InvalidArgument: code = "BAD_USER_INPUT"
   case codes.NotFound:        code = "NOT_FOUND"
   default:                    code = "INTERNAL" // log the details, don't return them
   }
   // &gqlerror.Error{Message: …, Extensions: map[string]any{"code": code}}
   ```
   You can call it from each resolver, or once in `srv.SetErrorPresenter`. Try both and decide which you prefer.
6. **HTTP.** Use `handler.New(graph.NewExecutableSchema(…))` with the `transport.POST{}` transport. Serve it on `POST /graphql`, `playground.Handler` on `GET /graphql`, and `GET /healthz`. Go 1.22+ `http.ServeMux` patterns like `"POST /graphql"` route on method for you.
7. **`Item.createdBy`.** Start with the simple version: a field resolver that calls `GetUser` for each item. To get one, map `Item` to your own model in `gqlgen.yml` that holds `CreatedByID string` instead of a `CreatedBy *User` field. gqlgen then can't find the field and generates an `Item().CreatedBy` resolver for you. It works, but it's slow, and you'll fix it in milestone 9.

**Think about it:** a page of 25 items now makes 26 gRPC calls. Where can you see that happening? Add a `grpc.WithUnaryInterceptor` to both clients that logs every call with `slog`, and look.

**Checkpoint:** in GraphiQL at `http://localhost:4000/graphql`:
```graphql
{
  items(offset: 0, limit: 5, orderBy: {field: QUANTITY, direction: DESC}, filter: {search: "cable"}) {
    totalCount
    nodes { id name quantity category }
  }
}
```
Also try `limit: 1000`. You should get `extensions.code = "BAD_USER_INPUT"`.

---

## Milestone 5: A Postgres database per service

**Goal:** replace the in-memory maps with a real database for each service, and move paging, sorting and filtering into SQL. See [data-model.drawio](data-model.drawio) and panel 4 of [graphql-api.drawio](graphql-api.drawio).

**Learn:** migrations, parameterised queries, building dynamic SQL **safely**, `COUNT(*) OVER ()`, `pg_trgm`, and why `ORDER BY` needs a tie-breaker.

### 5a. Run Postgres

Add two Postgres StatefulSets (`item-db`, `user-db`) in `infra/development/k8s/` with a Service and a Secret each, load them in the [Tiltfile](../Tiltfile), and port-forward (for example, item-db → `5433`, user-db → `5434`). Check with `psql`.

### 5b. Migrations
- Go: [goose](https://github.com/pressly/goose) with `//go:embed migrations/*.sql`, run at startup.
- Rust: [sqlx](https://github.com/launchbadge/sqlx) with `sqlx::migrate!()`, run at startup.
- The first migration creates the tables from the data-model diagram (leave `outbox` and `processed_events` for milestone 7) plus `CREATE EXTENSION IF NOT EXISTS pg_trgm;`.
- Add a dev-only seed migration that inserts the rows from `items.json`.

### 5c. The ListItems query

The key rule: **user input can only become bind parameters (`$1`), never SQL text.**

- **Sorting:** a `match` from the proto enum to a hard-coded column name, always followed by `, id ASC`.
  ```rust
  let col = match order.field() {
      ItemSortField::Quantity => "quantity",
      ItemSortField::Notes    => "notes",  // think about NULLs → NULLS LAST
      /* … */
      _ => "name",
  };
  ```
- **Filtering:** use `sqlx::QueryBuilder` and add `AND …` plus `push_bind(value)` only for fields that are set.
- **Search and contains filters:** `ILIKE $n` with the value `%{escaped}%`. Escape `%`, `_` and `\` in the user's text first. What happens if you don't and someone searches for `100%`?
- **Total count:** add `COUNT(*) OVER () AS total_count` to the select. It returns the total before `LIMIT` is applied, in the same query.
  - Edge case: if `offset` is past the end, you get zero rows and therefore no count. Fall back to a separate `SELECT COUNT(*)` in that case.
- **Indexes:** a `GIN (name gin_trgm_ops)` index. Use `EXPLAIN ANALYZE` to check that it's used. On a small table Postgres may skip it, so insert 100k rows with `generate_series` to see the difference.

**Think about it:** why does paging with `ORDER BY category` alone produce duplicate or missing rows between pages, while `ORDER BY category, id` doesn't?

**Checkpoint:** every sort field in both directions, every filter, and paging past the end should give the same results as the old in-memory version. Write a small table-driven test that compares the two.

---

## Milestone 6: Wire up the web app

**Goal:** the Home table talks to the real API, and [Home.tsx](../web/src/pages/Home.tsx) doesn't change at all.

The mock in [web/src/api/items.ts](../web/src/api/items.ts) was written for exactly this: each function's signature is the contract.

1. Write a `gqlRequest<T>(query, variables)` helper that sends a `fetch` to `GATEWAY_URL`, throws on `errors[]`, and returns `data`.
2. Replace the body of each mock function with a GraphQL call. The operation texts are already in the comment at the top of the file.
3. Change `filter.quantity` to an `IntFilter` and parse the column text in `toItemsVariables` (`"5"`, `">=5"`, `"<=5"`, `"1-10"`).
4. Delete the mock data and the in-memory store.

**Think about it:** SolidStart can call the gateway from the server (SSR, where `http://graphql-gateway:4000` resolves inside the cluster) or from the browser (which needs a port-forward or an ingress, plus CORS). Which do you want, and why?

**Checkpoint:** the table pages, sorts, searches and filters, and adding, editing and deleting items all work, with the same behaviour as the mock.

---

## Milestone 7: Kafka, the outbox and projections

**Goal:** services react to each other's changes **without calling each other**. See [event-flow.drawio](event-flow.drawio).

**Learn:** topics, partitions, keys, consumer groups, offsets, at-least-once delivery, the [transactional outbox pattern](https://microservices.io/patterns/data/transactional-outbox.html), and idempotent consumers.

### 7a. Run Kafka
A single-node Kafka in KRaft mode (the `apache/kafka` image) as a StatefulSet, plus a Job that creates the `user.events` and `item.events` topics. Port-forward it and check it with `kcat -L -b localhost:9092`.

### 7b. Why the outbox?
First try the naive version: in `DeleteUser`, commit the delete to the database and then produce to Kafka. Now ask yourself what happens if the process crashes between those two steps, and what happens if Kafka is down.

The fix:
1. In **one** database transaction, change the data **and** `INSERT INTO outbox (event_id, topic, aggregate_id, payload)`.
2. A background **relay** loop runs `SELECT … FROM outbox WHERE published_at IS NULL ORDER BY id LIMIT 100 FOR UPDATE SKIP LOCKED`, produces each row with `key = aggregate_id`, waits for the ack, then sets `published_at = now()`.
3. A crash after producing but before the `UPDATE` means the event is sent again. That's fine, because the consumers deduplicate.

Libraries: [rust-rdkafka](https://github.com/fede1024/rust-rdkafka) (`FutureProducer`, `StreamConsumer`) and [franz-go](https://github.com/twmb/franz-go).

### 7c. Consumers and projections
- **item-service** consumes `user.events` in the consumer group `item-service`:
  - `UserCreated` / `UserUpdated` → upsert into `users_projection`. Ignore events older than the stored `version`.
  - `UserDeleted` → `UPDATE items SET created_by = NULL …` and delete from the projection.
- **user-service** consumes `item.events` → `user_item_counts ± 1`.
- **Idempotency:** the first statement in the consumer's transaction is `INSERT INTO processed_events (event_id) … ON CONFLICT DO NOTHING`. If it inserted 0 rows, you've seen this event before, so skip it.
- Commit the Kafka offset **after** the database transaction commits, never before. Why?

**Think about it:**
- Why use `key = user_id`? What ordering guarantee does that give you, and what would break without it?
- Why does the event carry the data (a snapshot), rather than the consumer calling `GetUser` when it receives the event?

**Checkpoints:**
1. Create a user, create an item as that user, and look at `users_projection` in item-db. The user is there.
2. Delete the user. Within a second, the item's `created_by` becomes `NULL`.
3. Stop Kafka, create an item, and look at the outbox: the row is pending. Start Kafka again and it gets published.
4. Reset the consumer group offset to the beginning and replay. Nothing changes a second time.

---

## Milestone 8: Containers, k8s and Tilt

**Goal:** `tilt up` starts everything. Follow the existing `web` setup in the [Tiltfile](../Tiltfile) and [infra/development/](../infra/development/).

- **Dockerfiles** (build context = repo root, so `proto/` is available):
  - Rust: a multi-stage build using [cargo-chef](https://github.com/LukeMathWalker/cargo-chef) to cache dependencies. The build stage needs `protobuf-compiler` and `cmake`, and the runtime stage is `debian:bookworm-slim`.
  - Go: a `golang` build stage with `CGO_ENABLED=0`, and `gcr.io/distroless/static` at runtime.
- **k8s:** one Deployment and Service per service. gRPC ports are named `grpc`, and gRPC readiness probes use `grpc: { port: 50051 }`, which works because you registered the health service. Add addresses and broker lists to `app-config.yaml`, and put passwords in Secrets.
- **Tilt:** `docker_build` + `k8s_resource` with `labels` (`infra`, `backend`, `gateway`) and `resource_deps`, so services start after their database and Kafka.

**Checkpoint:** `tilt down && tilt up` from scratch goes all green, and milestone 6's checkpoint passes against the cluster.

---

## Milestone 9: Stretch goals

Pick whichever interest you:

- **DataLoader:** build a `UserLoader` with [dataloadgen](https://github.com/vikstrous/dataloadgen) whose fetch function calls `GetUsers`. Create a new loader **per request** in HTTP middleware and put it in the `context`. Why not one global loader? A page of 25 items should now make **2** gRPC calls instead of 26. Prove it with your interceptor logs. See gqlgen's [dataloaders recipe](https://gqlgen.com/reference/dataloaders/).
- **Auth propagation:** read the `Authorization` header in HTTP middleware, store it in the `context`, and copy it into outgoing gRPC metadata with a client interceptor (`metadata.AppendToOutgoingContext`). Add a server interceptor in each service that checks it.
- **Tracing:** OpenTelemetry across gateway → service → Kafka → consumer, carrying the trace context through gRPC metadata and Kafka headers.
- **Contract safety:** run `buf breaking --against '.git#branch=main'` in CI. Rename a field and watch it fail.
- **Tests:** integration tests against a real Postgres with testcontainers, one per filter or sort combination.
- **A third language:** add a small `location-service` in Python or TypeScript that uses the same proto files. That's when the polyglot promise is really tested.
- **Cursor pagination:** add a `itemsAfter(cursor, limit)` query that pages by `(sort_value, id)`, and compare its performance with `OFFSET` on 1M rows.

---

## Debugging cheat sheet

| Symptom | Likely cause |
|---|---|
| `grpcurl: server does not support the reflection API` | Reflection isn't registered |
| Gateway: `Unavailable` / `connection refused` | Wrong `*_SERVICE_ADDR`, or it has a scheme. grpc-go wants `host:port`, not `http://host:port` |
| Gateway: resolver changes vanish or won't compile | You edited `generated.go`, or forgot to re-run `gqlgen generate` after changing the schema |
| Rust build: `protoc not found` | Install `protobuf-compiler` or set `PROTOC` |
| Rust build fails in `rdkafka-sys` | `cmake` is missing |
| Duplicate or missing rows between pages | No `id` tie-breaker in `ORDER BY` |
| Search for `_` matches everything | `ILIKE` wildcards weren't escaped |
| Events processed twice | The offset was committed before the DB transaction, or there's no `processed_events` check |
| Consumer sees nothing | Wrong group, `auto.offset.reset` is `latest`, or the topic name doesn't match |

## Glossary

- **Gateway / BFF:** one API shaped for the frontend, sitting in front of many services.
- **Protobuf / gRPC:** a binary schema format, and the RPC framework built on it over HTTP/2.
- **Resolver:** the function that produces the value of one GraphQL field.
- **DataLoader:** batches and caches many single lookups made during one request into one call.
- **Outbox:** a table used to publish events reliably, written in the same transaction as the change.
- **Projection / read model:** a local copy of another service's data, kept up to date by events.
- **Idempotent consumer:** a consumer where processing the same event twice has the same effect as processing it once.
- **Consumer group:** consumers that share the work of reading a topic. Each partition goes to one member.
