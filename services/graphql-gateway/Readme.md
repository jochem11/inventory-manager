# graphql-gateway

The **single API for the web app**: one GraphQL endpoint in front of the gRPC services. On every request it:
- verifies the access token
- enforces permissions
- validates input
- calls the right services, translating their types and errors to GraphQL

It has no database and no business logic of its own.

| | |
|---|---|
| **API** | GraphQL on `:4000/graphql`, schema in [schema/](schema) |
| **Calls** | [auth-service](../auth-service/Readme.md) `:50052`, [user-service](../user-service/Readme.md) `:50051` |
| **Playground** | http://localhost:4000/graphql in a browser |

![Overview](docs/overview.svg)

## Routes

| Route | |
|---|---|
| `POST /graphql` | GraphQL queries and mutations |
| `GET /graphql` | GraphiQL playground |
| `GET /healthz` | Health check |

Requests pass three middlewares, in this order:
1. **Tracing:** one OpenTelemetry span per request, continued into the services.
2. **CORS:** only the origins in `ALLOWED_ORIGINS` may call, with credentials, so the browser can send and receive the refresh cookie.
3. **Authentication:** verifies `Authorization: Bearer <token>` and puts its claims in the request context. A missing or expired token isn't an error by itself; fields that need a login answer `UNAUTHENTICATED`.

## API

| Operation | Requires | Backed by |
|---|---|---|
| `me` | login | user-service `GetUser` |
| `user(id)`, `users(offset, limit, orderBy, filter)` | `users:read` | user-service `GetUser`, `ListUsers` |
| `createUser`, `updateUser`, `deleteUser` | `users:write` | user-service |
| `register`, `verifyEmail`, `resendVerification` | | auth-service |
| `login`, `refreshToken`, `logout` | the refresh cookie (for refresh and logout) | auth-service |

Lists share one shape: `things(offset, limit, orderBy, filter): ThingConnection!`, with `totalCount` and `nodes`. Offset ≥ 0, limit 1–100.

`login` and `refreshToken` return an `AuthPayload` with:
- the access token and its expiry
- the user's `roles` and `permissions`, so clients can show only what's allowed
- the user's profile

## Access control

Fields are protected in the schema with directives, which run before the resolver:

```graphql
me: User @authenticated
users(...): UserConnection! @hasPermission(permission: "users:read")
reports: [Report!]! @hasRole(roles: ["admin"])
```

| Situation | `extensions.code` | HTTP status |
|---|---|---|
| No valid access token | `UNAUTHENTICATED` | 200 |
| Missing permission or role | `FORBIDDEN` | 403 |

- **Code:** the directives are in [internal/auth/directives.go](internal/auth/directives.go). In Go code, `auth.RequirePermission(ctx, …)` and `auth.RequireRole(ctx, …)` do the same checks.
- **Prefer permissions to roles:** a role is only a named set of permissions, defined in [shared/auth/permissions.go](../../shared/auth/permissions.go).

### Token verification

![Authenticated request](docs/authenticated-request.svg)

The gateway verifies access tokens itself, so an ordinary request never calls the auth-service.

**Keys:** the auth-service's public keys are cached by key id (`kid`). An unknown `kid` triggers a refetch through `GetPublicKeys`, at most once every 10 s.

**A token is accepted only with:**
- a valid EdDSA signature
- an expiry in the future
- issuer `auth-service`
- audience `inventory-manager`

### Sessions and the refresh cookie

![Login and refresh cookie](docs/login-refresh.svg)

The refresh token never appears in a GraphQL response: the gateway keeps it in a cookie.

**The cookie:**
- `HttpOnly`: scripts can't read it
- `Path=/graphql`: only sent to this endpoint
- `SameSite=Lax`: not sent with requests started from other sites
- expires with the 30-day session

**Lifecycle:**
- `refreshToken` swaps the cookie for a new one.
- An expired session clears the cookie.
- `logout` ends the session and clears the cookie.

## Errors

Service errors arrive as gRPC status codes, and [internal/gqlerr](internal/gqlerr) translates them into `extensions.code`:

| Code | From | Meaning |
|---|---|---|
| `UNAUTHENTICATED` | no valid token; gRPC `UNAUTHENTICATED` | Log in, or refresh the token |
| `FORBIDDEN` | missing permission or role | HTTP 403 |
| `BAD_USER_INPUT` | gRPC `INVALID_ARGUMENT` | `extensions.fields` holds a message per field |
| `NOT_FOUND` | gRPC `NOT_FOUND` | |
| `CONFLICT` | gRPC `ALREADY_EXISTS` | e.g. an email that's taken |
| `FAILED_PRECONDITION` | gRPC `FAILED_PRECONDITION` | e.g. logging in before verifying the email |
| `INTERNAL` | anything else | The cause is logged, not returned |

## Configuration

| Variable | Default | Notes |
|---|---|---|
| `HTTP_ADDR` | `:4000` | |
| `USER_SERVICE_ADDR` | `localhost:50051` | `host:port`, no scheme |
| `AUTH_SERVICE_ADDR` | `localhost:50052` | |
| `ALLOWED_ORIGINS` | `http://localhost:3000` | Comma-separated origins allowed to call with cookies |
| `COOKIE_SECURE` | `false` | Set `true` behind HTTPS |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | unset (no tracing) | In the cluster: `http://jaeger:4317` |

## Development

```sh
make run       # start the gateway on :4000
make graphql   # regenerate the GraphQL code after changing schema/*.graphqls
make proto     # regenerate the gRPC clients after changing proto/
make           # list all targets
```

In the playground, put a token in the *Headers* tab: `{"Authorization": "Bearer …"}`.

### Adding a service

Using an `item` service as the example:

1. **Proto:** add `item/item.proto` to the [Makefile](Makefile), the same way as `user` and `auth`, and run `make proto`.
2. **Schema:** add `schema/item.graphqls` with `extend type Query` / `extend type Mutation`, and guard each field with `@hasPermission`. Run `make graphql`; it adds `internal/graph/item.resolvers.go` with stubs.
3. **Resolver:** add `internal/item/` with a `Resolver` around the gRPC client and a `mapper.go`, like `internal/user/`. Lists use `list.ValidatePage`. Return gRPC errors unchanged; `gqlerr` translates them.
4. **Wire it:** add `ItemResolver` to [internal/graph/resolver.go](internal/graph/resolver.go), make each stub a one-line call into it, and dial the service in [cmd/main.go](cmd/main.go) using an `ITEM_SERVICE_ADDR` variable.
5. **Deploy:**
   - add `ITEM_SERVICE_ADDR` to the [deployment](../../infra/development/k8s/graphql-gateway-deployment.yaml)
   - add the service to the gateway's `resource_deps` in the [Tiltfile](../../Tiltfile)

## Project structure

```
cmd/main.go                  startup: gRPC clients, resolvers, middleware, routes
schema/                      GraphQL schema, one .graphqls per service
internal/
  auth/                      token verification, directives, refresh cookie, auth resolver
  user/                      user resolver (calls the user-service) and mapping
  gqlerr/                    error codes, gRPC status → GraphQL error, HTTP 403
  list/                      paging validation, sort direction
  clients/                   gRPC dialing and call logging
  graph/                     gqlgen glue: Resolver, server, one-line resolvers
    generated/, model/       generated by gqlgen, don't edit
pkg/pb/                      generated from proto/, don't edit
docs/                        graphql-gateway.drawio and its SVGs
```
