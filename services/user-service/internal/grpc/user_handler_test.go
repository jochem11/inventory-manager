package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/jochem11/inventory-manager/shared/paging"
	"net"
	"testing"
	"time"

	"github.com/jochem11/inventory-manager/services/user-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/user-service/internal/models"
	userpb "github.com/jochem11/inventory-manager/services/user-service/pkg/pb/user"
	"github.com/jochem11/inventory-manager/services/user-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/database"
	"github.com/jochem11/inventory-manager/shared/errs"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// fakeService is a domain.UserService that records what the handler passed
// in and returns err (when set) or a fixed user.
type fakeService struct {
	err        error
	input      types.UserInput
	listParams types.UserListParams
	id         string
	ids        []string
}

var ada = &models.User{
	Model: database.Model{
		ID:        "3K4C4qfTOu1pNlLy3ZDPhidNLBr",
		CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC),
	},
	FirstName: "Ada",
	LastName:  "Lovelace",
	Email:     "ada@example.com",
	Phone:     ptr("+31612345678"),
}

func ptr(s string) *string { return &s }

func (f *fakeService) result() (*models.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	return ada, nil
}

func (f *fakeService) Create(_ context.Context, input types.UserInput) (*models.User, error) {
	f.input = input
	return f.result()
}

func (f *fakeService) CreateWithID(_ context.Context, id string, input types.UserInput) (*models.User, error) {
	f.id, f.input = id, input
	return f.result()
}

func (f *fakeService) FindByID(_ context.Context, id string) (*models.User, error) {
	f.id = id
	return f.result()
}

func (f *fakeService) FindByEmail(_ context.Context, _ string) (*models.User, error) {
	return f.result()
}

func (f *fakeService) FindByIDs(_ context.Context, ids []string) ([]*models.User, error) {
	f.ids = ids
	if f.err != nil {
		return nil, f.err
	}
	return []*models.User{ada}, nil
}

func (f *fakeService) FindAll(_ context.Context, params types.UserListParams) (*types.Page[*models.User], error) {
	f.listParams = params
	if f.err != nil {
		return nil, f.err
	}
	return &types.Page[*models.User]{TotalCount: 42, Nodes: []*models.User{ada}}, nil
}

func (f *fakeService) Update(_ context.Context, id string, input types.UserInput) (*models.User, error) {
	f.id, f.input = id, input
	return f.result()
}

func (f *fakeService) Delete(_ context.Context, id string) error {
	f.id = id
	return f.err
}

// newClient serves the handler over an in-memory connection, so requests go
// through real gRPC encoding and status handling without a network port.
func newClient(t *testing.T, svc domain.UserService) userpb.UserServiceClient {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	server := grpc.NewServer()
	userpb.RegisterUserServiceServer(server, NewUserHandler(svc))
	go server.Serve(lis)
	t.Cleanup(server.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return userpb.NewUserServiceClient(conn)
}

func TestGetUserConvertsUser(t *testing.T) {
	svc := &fakeService{}
	resp, err := newClient(t, svc).GetUser(context.Background(), &userpb.GetUserRequest{Id: ada.ID})
	if err != nil {
		t.Fatal(err)
	}

	u := resp.GetUser()
	if svc.id != ada.ID {
		t.Errorf("service got id %q, want %q", svc.id, ada.ID)
	}
	if u.GetId() != ada.ID || u.GetFirstName() != "Ada" || u.GetLastName() != "Lovelace" || u.GetEmail() != "ada@example.com" {
		t.Errorf("user fields not converted: %v", u)
	}
	if u.GetPhone() != "+31612345678" || u.AvatarUrl != nil {
		t.Errorf("optional fields: phone %v, avatar %v; want phone set and avatar unset", u.Phone, u.AvatarUrl)
	}
	if !u.GetCreatedAt().AsTime().Equal(ada.CreatedAt) || !u.GetUpdatedAt().AsTime().Equal(ada.UpdatedAt) {
		t.Errorf("timestamps: %v, %v", u.GetCreatedAt(), u.GetUpdatedAt())
	}
}

func TestCreateAndUpdatePassInput(t *testing.T) {
	svc := &fakeService{}
	client := newClient(t, svc)
	input := &userpb.UserInput{FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com", AvatarUrl: ptr("https://example.com/a.png")}

	if _, err := client.CreateUser(context.Background(), &userpb.CreateUserRequest{User: input}); err != nil {
		t.Fatal(err)
	}
	want := types.UserInput{FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.com"}
	if got := svc.input; got.FirstName != want.FirstName || got.Email != want.Email || got.Phone != nil || got.AvatarURL == nil || *got.AvatarURL != "https://example.com/a.png" {
		t.Errorf("CreateUser passed %+v", got)
	}

	if _, err := client.UpdateUser(context.Background(), &userpb.UpdateUserRequest{Id: ada.ID}); err != nil {
		t.Fatal(err)
	}
	if svc.id != ada.ID || svc.input != (types.UserInput{}) {
		t.Errorf("UpdateUser without a user should pass empty input for id %s, got %q %+v", ada.ID, svc.id, svc.input)
	}
}

func TestGetUsersPassesIDs(t *testing.T) {
	svc := &fakeService{}
	resp, err := newClient(t, svc).GetUsers(context.Background(), &userpb.GetUsersRequest{Ids: []string{"a", "b"}})
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprint(svc.ids) != "[a b]" || len(resp.GetUsers()) != 1 {
		t.Errorf("ids %v, users %d", svc.ids, len(resp.GetUsers()))
	}
}

func TestListUsersMapsRequest(t *testing.T) {
	tests := []struct {
		name string
		req  *userpb.ListUsersRequest
		want types.UserListParams
	}{
		{
			name: "empty request",
			req:  &userpb.ListUsersRequest{},
			want: types.UserListParams{},
		},
		{
			name: "everything set",
			req: &userpb.ListUsersRequest{
				Offset:        20,
				Limit:         10,
				OrderBy:       "last_name",
				SortDirection: "desc",
				Filter: &userpb.UserFilter{
					Search: ptr("jan"), FirstName: ptr("li"), LastName: ptr("jansen"), Email: ptr("demo"), Phone: ptr("+31"),
				},
			},
			want: types.UserListParams{
				Params:  paging.Params{Offset: 20, Limit: 10},
				OrderBy: types.UserOrder{Field: types.UserSortLastName, Direction: "desc"},
				Filter:  types.UserFilter{Search: "jan", FirstName: "li", LastName: "jansen", Email: "demo", Phone: "+31"},
			},
		},
		{
			name: "direction defaults to ascending",
			req:  &userpb.ListUsersRequest{OrderBy: "email"},
			want: types.UserListParams{OrderBy: types.UserOrder{Field: types.UserSortEmail, Direction: types.SortAsc}},
		},
		{
			name: "direction without order_by is ignored",
			req:  &userpb.ListUsersRequest{SortDirection: "desc"},
			want: types.UserListParams{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeService{}
			resp, err := newClient(t, svc).ListUsers(context.Background(), tt.req)
			if err != nil {
				t.Fatal(err)
			}

			got := svc.listParams
			if got != tt.want {
				t.Errorf("params = %+v, want %+v", got, tt.want)
			}
			if resp.GetTotalCount() != 42 || len(resp.GetUsers()) != 1 {
				t.Errorf("response: total %d, %d users", resp.GetTotalCount(), len(resp.GetUsers()))
			}
		})
	}
}

func TestErrorsBecomeStatusCodes(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantCode    codes.Code
		wantMessage string
	}{
		{"not found", fmt.Errorf("find user: %w", domain.ErrUserNotFound), codes.NotFound, "user not found"},
		{"email taken", fmt.Errorf(`"a@b.c": %w`, domain.ErrEmailTaken), codes.AlreadyExists, "email is already in use"},
		{"canceled", context.Canceled, codes.Canceled, "context canceled"},
		{"deadline", fmt.Errorf("query: %w", context.DeadlineExceeded), codes.DeadlineExceeded, "query: context deadline exceeded"},
		// Unexpected errors mustn't leak internals to clients.
		{"unexpected", errors.New("dial tcp 10.0.0.5:3306: connection refused"), codes.Internal, "internal error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newClient(t, &fakeService{err: tt.err}).GetUser(context.Background(), &userpb.GetUserRequest{Id: "x"})

			st := status.Convert(err)
			if st.Code() != tt.wantCode || st.Message() != tt.wantMessage {
				t.Errorf("status = %v %q, want %v %q", st.Code(), st.Message(), tt.wantCode, tt.wantMessage)
			}
		})
	}
}

func TestValidationErrorHasFieldDetails(t *testing.T) {
	invalid := &errs.ValidationError{Fields: map[string]string{
		"firstName": "firstName is a required field",
		"email":     "email must be a valid email address",
	}}
	_, err := newClient(t, &fakeService{err: invalid}).CreateUser(context.Background(), &userpb.CreateUserRequest{})

	st := status.Convert(err)
	if st.Code() != codes.InvalidArgument {
		t.Fatalf("code = %v, want InvalidArgument", st.Code())
	}
	if len(st.Details()) != 1 {
		t.Fatalf("details = %v, want one BadRequest", st.Details())
	}
	badRequest, ok := st.Details()[0].(*errdetails.BadRequest)
	if !ok {
		t.Fatalf("detail is %T, want *errdetails.BadRequest", st.Details()[0])
	}

	var got []string
	for _, v := range badRequest.GetFieldViolations() {
		got = append(got, v.GetField()+": "+v.GetDescription())
	}
	// Sorted by field, so clients get a stable order.
	want := []string{"email: email must be a valid email address", "firstName: firstName is a required field"}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("violations = %q, want %q", got, want)
	}
}

func TestDeleteUser(t *testing.T) {
	svc := &fakeService{}
	if _, err := newClient(t, svc).DeleteUser(context.Background(), &userpb.DeleteUserRequest{Id: ada.ID}); err != nil {
		t.Fatal(err)
	}
	if svc.id != ada.ID {
		t.Errorf("service got id %q", svc.id)
	}
}
