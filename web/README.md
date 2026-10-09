# web

The **inventory-manager app**, built with [SolidStart](https://start.solidjs.com). Here users:
- sign up and log in
- browse and manage items
- manage users (admins only)
- see locations on a floor plan

The server renders the page shell; data loads in the browser from the [graphql-gateway](../services/graphql-gateway/Readme.md).

| | |
|---|---|
| **URL** | http://localhost:3000 |
| **Talks to** | the graphql-gateway, `VITE_GRAPHQL_URL` |
| **Stack** | SolidStart, TypeScript, Zod, Sass, Leaflet |

![Overview](docs/overview.svg)

## Pages

| Path | Page | Access | Data |
|---|---|---|---|
| `/login`, `/register`, `/verify-email` | Sign-in and sign-up | public | auth mutations |
| `/` | All items: search, filter by category or status, sort, add, edit, duplicate, delete | `items:read`; changes need `items:write` | `items`, `createItem`, `updateItem`, `deleteItem` |
| `/categories`, `/statuses` | Categories and statuses: search, sort, add, rename, delete | `items:read`; changes need `items:write` | `categories`, `itemStatuses` and their mutations |
| `/users` | Users: search, sort, filter, edit, delete | `users:read`; editing needs `users:write` | `users`, `updateUser`, `deleteUser` |
| `/locations` | Floor plan with locations (Leaflet, GeoJSON) | login | [data/floorPlan.json](src/data/floorPlan.json) |
| `/reports`, `/settings/*` | Coming soon | login | |

## Authentication

### Restoring the session

![Opening the app](docs/open-app.svg)

**Where the tokens live:**
- **Access token:** only in memory, so it can't be read from storage.
- **Refresh token:** in an httpOnly cookie that the gateway manages.

On every page load, `AuthProvider` exchanges the cookie for a fresh access token. While that's in progress, protected pages render nothing. Without a session, they redirect to `/login?redirect=…`.

### Logging in

![Logging in](docs/login.svg)

- **Validation:** forms are checked with Zod in the browser; the services check again.
- **Redirect:** after login, the user returns to the page they came from, but only to a path on this site.
- **Unverified accounts** can request a new activation link.

### Requests and token refresh

![Requests and refresh](docs/request.svg)

`useAuth().request(query, variables)` sends GraphQL requests as the logged-in user.

**When it refreshes the access token:**
- 60 seconds before it expires
- once, when a request comes back `UNAUTHENTICATED`; the request is then retried

**Concurrent refreshes share one request.** That matters because the refresh cookie rotates on every use.

### Permissions in the UI

`login` and `refreshToken` return the user's `roles` and `permissions`. The catalog is in [constants/access.ts](src/constants/access.ts).

| Tool | Use |
|---|---|
| `useAuth().hasPermission("users:write")`, `hasRole("admin")`, `allows({ permission, role })` | Checks in code |
| `<Guard permission="users:read" fallback={<Forbidden />}>…</Guard>` | Show content only with access |
| `access: { permission: "users:read" }` on a nav item | Hide the item from the sidebar |

These checks only shape the UI; the gateway enforces every request.

## Data tables

![Users table](docs/users-table.svg)

Tables page, sort, filter and search on the server. Every list in the API has the same shape, so a new table needs four pieces:

1. **An API module** with the type, a `fetchThings(request, variables: ListVariables)`, and a sort map from column id to the API's sort field. Column ids that match the API's filter fields need no further mapping.
   ```ts
   export const THING_SORT_FIELDS = { name: "NAME", createdAt: "CREATED_AT" };
   ```
2. **Columns:** a `DataTableColumn<Thing>[]`.
3. **The data source:**
   ```ts
   const things = createServerTable({
     fetch: (variables) => api.fetchThings(auth.request, variables),
     sortFields: api.THING_SORT_FIELDS,
   });
   ```
4. **The table:**
   ```tsx
   <PageHeader title="Things" description={things.summary("thing", "things")} />
   <DataTable source={things} columns={THING_COLUMNS} rowKey={(t) => t.id} />
   ```

**Included:**
- the search is kept in the URL (`?q=`), and the topbar search fills it
- searching is debounced
- out-of-order responses are ignored
- load errors show in the table with a retry

**Changes:**
- `createDeleteAction({ source, noun, name, remove })` deletes with a confirmation.
- `createFormDialog().open({ title, schema, fields, onSubmit })` edits in a dialog.

Both refresh the table afterwards. The [users](src/features/users) and [items](src/features/items) features use all of these.

## Forms

- **Field rules:** [schemas/fields.ts](src/schemas/fields.ts) holds the rules the services also check: `nameField`, `emailField`, `newPasswordField`, `optionalPhoneField`. Compose schemas from them, so client and server agree.
- **Server errors:** `fieldErrorMessage(error)` from `api/graphql.ts` turns them into one message, listing every invalid field for `BAD_USER_INPUT`.
- **Notices:** `<Notice tone="error" | "success" | "info">` shows them, and `<ErrorNotice message={…} />` when the message may be empty.

## Configuration

| Variable | Default |
|---|---|
| `VITE_GRAPHQL_URL` | `http://localhost:4000/graphql` |

The gateway only accepts cookies from the origins in its `ALLOWED_ORIGINS`. Add the app's address there when it runs somewhere other than `http://localhost:3000`.

## Development

In Tilt, changes under `src/` reach the running app without a rebuild. To run it on its own:

```sh
bun install
bun run dev      # http://localhost:3000
bun run build    # production build in .output/; start it with `bun run start`
```

**Adding a page:**
1. Create `src/pages/Thing.tsx`.
2. Register it in [routes.ts](src/routes.ts).
3. Add it to [navigation.ts](src/constants/navigation.ts), with `access` when it needs a permission.

## Project structure

```
src/
  app.tsx, routes.ts     providers, router, layouts
  api/                   GraphQL operations per area; list.ts holds the list convention
  providers/             AuthProvider, ThemeProvider, DialogProvider
  context/               contexts and hooks: useAuth(), useTheme(), …
  pages/                 one component per route
  features/              larger parts with their own components and hooks: auth/, layout/, users/, items/,
                         namedRecords/ (one page for categories and statuses)
  schemas/               Zod field rules shared by the forms
  components/            reusable UI: ui/ (DataTable, Button, …), form/, dialogs/
  hooks/                 createServerTable, createDeleteAction, createFormDialog, createForm, …
  constants/             navigation, access (roles/permissions), API URL, theme
docs/                    web.drawio and its SVGs
```
