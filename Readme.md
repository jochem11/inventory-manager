# inventory-manager

A learning project: an inventory app built as small services.

| Part | What | README |
|---|---|---|
| [web](web) | SolidStart front-end | [web/README.md](web/README.md) |
| [graphql-gateway](services/graphql-gateway) | The one GraphQL endpoint; checks logins and permissions | [Readme](services/graphql-gateway/Readme.md) |
| [auth-service](services/auth-service) | Logins, email verification, sessions, JWTs, roles | [Readme](services/auth-service/Readme.md) |
| [user-service](services/user-service) | User profiles | [Readme](services/user-service/Readme.md) |
| [item-service](services/item-service) | The inventory (work in progress) | [Readme](services/item-service/Readme.md) |

The services talk gRPC, and Kafka for events. MySQL runs on your machine. Jaeger shows the traces.

## Running it

**You need:**
- [Go](https://go.dev) 1.26
- [Bun](https://bun.sh)
- Docker
- [minikube](https://minikube.sigs.k8s.io)
- [Tilt](https://tilt.dev)
- MySQL 8 on your machine (port 3306)

**Steps:**

1. **Settings.** Copy the example settings and fill in your MySQL user and password:
   ```sh
   cp .env.example .env
   ```
   `.env` holds credentials and is never committed. [.env.example](.env.example) explains each value, including how to generate `JWT_PRIVATE_KEY`.
2. **Start everything:**
   ```sh
   minikube start --cpus=3
   tilt up
   ```
   Tilt compiles the Go services on your machine, builds the images, and deploys everything to minikube.
3. **Open the app:**

   | What | Where |
   |---|---|
   | App | http://localhost:3000 |
   | GraphQL playground | http://localhost:4000/graphql |
   | Jaeger (traces) | http://localhost:16686 |
   | Kafka UI (start it in Tilt) | http://localhost:8080 |
4. **Create an account** at `/register`. Email sending isn't built yet: open the auth-service logs in Tilt and click the activation link.
5. **Become admin** to see the admin pages, like Users: put your email in `ADMIN_EMAILS` in `.env` and restart the auth-service in Tilt.

## More

- [docs/LEARNING_GUIDE.md](docs/LEARNING_GUIDE.md): the milestones this project follows.
- [docs/](docs): draw.io diagrams of the architecture and the auth flows.
