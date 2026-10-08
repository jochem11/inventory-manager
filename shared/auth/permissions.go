package auth

// The built-in roles. Every identity gets RoleUser at registration; RoleAdmin
// is given to the emails in the auth-service's ADMIN_EMAILS.
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// Permissions are "resource:action". Checks should prefer permissions over
// roles: a role is just a named set of permissions, so permissions keep
// working when roles are added or regrouped.
const (
	PermUsersRead  = "users:read"
	PermUsersWrite = "users:write"
	PermItemsRead  = "items:read"
	PermItemsWrite = "items:write"
)

// RolePermissions is what each role grants. The auth-service makes the
// database match this at every start, so a change here takes effect after a
// restart, and in a user's token after their next refresh.
var RolePermissions = map[string][]string{
	RoleUser:  {PermItemsRead, PermItemsWrite},
	RoleAdmin: {PermItemsRead, PermItemsWrite, PermUsersRead, PermUsersWrite},
}
