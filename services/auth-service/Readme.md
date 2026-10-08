# auth-service

Owns **logins**: passwords, email verification, sessions, and the signed access tokens the [graphql-gateway](../graphql-gateway/Readme.md) checks on every request. It also holds the roles and permissions that go into those tokens.

| | |
|---|---|
| **API** | gRPC `auth.AuthService` on `:50052`, defined in [proto/auth/auth.proto](../../proto/auth/auth.proto) |
| **Database** | MySQL `auth_service`, created on first start |
| **Publishes** | `auth.events`: `IdentityRegistered` |
| **Consumes** | `user.events`: `UserDeleted` |

Profiles (name, phone, avatar) belong to the [user-service](../user-service/Readme.md). Both services use the same user id: the auth-service creates it at registration and announces it with `IdentityRegistered`.

![Overview](docs/overview.svg)

## API

| RPC | Description | Errors |
|---|---|---|
| `Register(email, password, names, phone)` | Creates the login, sends the activation link, publishes `IdentityRegistered` | `INVALID_ARGUMENT`, `ALREADY_EXISTS` |
| `VerifyEmail(token)` | Activates the account | `UNAUTHENTICATED` |
| `ResendVerification(email)` | Sends a new activation link; always succeeds | |
| `Login(email, password)` | Starts a session: access token + refresh token | `UNAUTHENTICATED`, `FAILED_PRECONDITION` (not verified) |
| `Refresh(refresh_token)` | A new access token, and replaces the refresh token | `UNAUTHENTICATED` |
| `Logout(refresh_token)` | Ends the session; always succeeds | |
| `GetPublicKeys()` | The public keys that verify access tokens | |

> Email delivery isn't wired up yet: activation links are written to the service log.

## Events

| | Topic | Event | Effect |
|---|---|---|---|
| Publishes | `auth.events` | `IdentityRegistered` | The user-service creates the profile |
| Consumes | `user.events` | `UserDeleted` | Deletes the login, its sessions and tokens |

Events go out through a transactional outbox. The consumer skips events it has already handled.

## Tokens

| | Access token | Refresh token | Activation token |
|---|---|---|---|
| **Format** | JWT, Ed25519 (EdDSA) | 32 random bytes | 32 random bytes |
| **Lifetime** | 15 minutes | 30 days, replaced on every use | 24 hours, single use |
| **Stored here** | no | SHA-256 only, in `sessions` | SHA-256 only, in `identity_tokens` |
| **Kept by the client** | in memory (web app) | httpOnly cookie, set by the gateway | in the link |

**Access token claims:**
- `sub`: the user id
- `sid`: the session id
- `email`
- `roles`
- `permissions`

The issuer is `auth-service` and the audience is `inventory-manager`. The header's `kid` tells verifiers which public key to use.

The gateway verifies access tokens locally, so normal requests never reach this service. Logging out or deleting a user stops further refreshes right away; an access token already issued stays valid until it expires, at most 15 minutes later.

## Roles and permissions

Defined once in [shared/auth/permissions.go](../../shared/auth/permissions.go):

| Role | Given to | Permissions |
|---|---|---|
| `user` | every account, at registration | `items:read`, `items:write` |
| `admin` | the emails in `ADMIN_EMAILS` | `items:read`, `items:write`, `users:read`, `users:write` |

At every start, [the seed](internal/repository/seed.go):
1. makes the database match that file, adding and removing permissions as needed
2. gives `user` to accounts without a role
3. gives `admin` to the emails in `ADMIN_EMAILS`

Changes reach a user with their next token refresh, within 15 minutes. The gateway enforces permissions with schema directives; see [Access control](../graphql-gateway/Readme.md#access-control).

## Workflows

### Register

![Register](docs/register.svg)

The identity, its role, the activation token's hash and the `IdentityRegistered` event are saved in one transaction.

### Verify email

![VerifyEmail](docs/verify-email.svg)

An activation link works once, for 24 hours, and only for the address it was sent to. `ResendVerification` always answers the same way, so it doesn't reveal which emails have an account.

### Login

![Login](docs/login.svg)

Unknown emails still go through a password check, against a dummy hash, so response times don't reveal which accounts exist. Accounts must be verified before they can log in.

### Refresh and logout

![Refresh and Logout](docs/refresh-logout.svg)

Every refresh replaces the refresh token: a copied token stops working as soon as the real owner refreshes. Logout revokes the session.

### When a user is deleted

![UserDeleted](docs/user-deleted.svg)

The identity is deleted, and its sessions and tokens go with it (`ON DELETE CASCADE`).

## Data model

![Data model](docs/data-model.svg)

## Configuration

| Variable | Default | Notes |
|---|---|---|
| `GRPC_ADDR` | `:50052` | |
| `DB_HOST`, `DB_PORT` | `localhost`, `3306` | |
| `DB_USER`, `DB_PASSWORD` | `root`, none | From `.env` |
| `DB_NAME` | `auth_service` | |
| `ADMIN_EMAILS` | none | From `.env`. Comma-separated accounts that get the `admin` role. |
| `JWT_PRIVATE_KEY` | a new key per start | From `.env`. Ed25519 PKCS#8 PEM on one line, with `\n` for line breaks. Without it, every restart logs everyone out. |
| `VERIFY_EMAIL_URL` | `http://localhost:3000/verify-email` | The page the activation link opens |
| `KAFKA_BROKERS` | `127.0.0.1:9094` | In the cluster: `kafka:9092` |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | unset (no tracing) | In the cluster: `http://jaeger:4317` |

To generate a signing key for `.env`:
```sh
openssl genpkey -algorithm ed25519 | awk 'NF {printf "%s\\n", $0}'
```

## Development

```sh
make run     # start the gRPC server on :50052 (reads ../../.env)
make proto   # regenerate pkg/pb after changing the .proto files
make         # list all targets
```

## Project structure

```
cmd/main.go            startup: tracing, database + seed, signing key, Kafka, gRPC
service/               AuthService: register, verify, login, refresh, logout
internal/
  domain/              interfaces, errors, token types
  grpc/                gRPC handler, errors → status codes
  repository/          GORM queries for identities, tokens, sessions; role seed
  token/               access-token signing (Ed25519)
  events/              Kafka consumer for user.events
  models/              Identity, IdentityToken, Session, Role, Permission
pkg/pb/                generated from proto/, don't edit
docs/                  auth-service.drawio and its SVGs
```
