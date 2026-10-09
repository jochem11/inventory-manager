package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jochem11/inventory-manager/shared/paging"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jochem11/inventory-manager/services/user-service/internal/domain"
	"github.com/jochem11/inventory-manager/services/user-service/internal/models"
	"github.com/jochem11/inventory-manager/services/user-service/pkg/types"
	"github.com/jochem11/inventory-manager/shared/database"
	"github.com/jochem11/inventory-manager/shared/kafka"
)

// These tests run against a real MySQL (DB_HOST etc., see
// database.ConfigFromEnv) in a separate database, which is reset per test.
// They are skipped when MySQL isn't reachable.
const testDBName = "user_service_test"

var fixtureNames = [][2]string{
	{"Emma", "de Vries"}, {"Liam", "Jansen"}, {"Sophie", "Bakker"}, {"Noah", "Visser"},
	{"Julia", "Smit"}, {"Daan", "Meijer"}, {"Mila", "de Boer"}, {"Sem", "Mulder"},
	{"Tess", "de Groot"}, {"Lucas", "Bos"}, {"Zoë", "Jansen"}, {"Eva", "O'Brien"},
	{"Nina", "Bakker"}, {"Tim", "Jansen"}, {"Max", "van den Berg"},
}

// fixtures returns 15 users: every third has a phone number, the Jansens and
// Bakkers tie when sorting by last name, and Sem's email has an underscore.
func fixtures() []*models.User {
	users := make([]*models.User, len(fixtureNames))
	for i, n := range fixtureNames {
		local := strings.ToLower(n[0] + "." + strings.NewReplacer(" ", "", "'", "").Replace(n[1]))
		if n[0] == "Sem" {
			local = strings.Replace(local, ".", "_", 1)
		}
		u := &models.User{FirstName: n[0], LastName: n[1], Email: local + "@example.com"}
		if i%3 == 0 {
			u.Phone = new(fmt.Sprintf("+316%08d", i))
		}
		users[i] = u
	}
	return users
}

func newTestRepo(t *testing.T) (domain.UserRepository, []*models.User) {
	t.Helper()
	cfg := database.ConfigFromEnv("user_service")

	server := cfg
	server.Name = ""
	admin, err := database.Connect(server)
	if err != nil {
		t.Skipf("MySQL not reachable: %v", err)
	}
	if err := admin.Exec("CREATE DATABASE IF NOT EXISTS " + testDBName).Error; err != nil {
		t.Fatalf("create test database: %v", err)
	}

	cfg.Name = testDBName
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	tables := append([]any{&models.User{}}, kafka.Models()...)
	if err := db.Migrator().DropTable(tables...); err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(tables...); err != nil {
		t.Fatal(err)
	}

	repo := NewUserRepository(db)
	users := fixtures()
	for _, u := range users {
		if err := repo.Create(context.Background(), u); err != nil {
			t.Fatalf("create %s: %v", u.Email, err)
		}
	}
	return repo, users
}

func emails(users []*models.User) []string {
	out := make([]string, len(users))
	for i, u := range users {
		out[i] = u.Email
	}
	return out
}

func sortedByID(users []*models.User) []*models.User {
	return slices.SortedFunc(slices.Values(users), func(a, b *models.User) int { return strings.Compare(a.ID, b.ID) })
}

func TestFindAllPaging(t *testing.T) {
	repo, all := newTestRepo(t)
	ctx := context.Background()

	var seen []*models.User
	for offset := 0; offset < len(all); offset += 4 {
		page, err := repo.FindAll(ctx, types.UserListParams{Params: paging.Params{Offset: offset, Limit: 4}})
		if err != nil {
			t.Fatal(err)
		}
		if page.TotalCount != int64(len(all)) {
			t.Fatalf("offset %d: total = %d, want %d", offset, page.TotalCount, len(all))
		}
		if want := min(4, len(all)-offset); len(page.Nodes) != want {
			t.Fatalf("offset %d: %d rows, want %d", offset, len(page.Nodes), want)
		}
		seen = append(seen, page.Nodes...)
	}
	// Without a sort, pages follow the ID, and together hold every user once.
	if got, want := emails(seen), emails(sortedByID(all)); !slices.Equal(got, want) {
		t.Fatalf("pages:\ngot  %v\nwant %v", got, want)
	}

	page, err := repo.FindAll(ctx, types.UserListParams{Params: paging.Params{Offset: 100, Limit: 10}})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Nodes) != 0 || page.TotalCount != int64(len(all)) {
		t.Errorf("past the end: %d rows, total %d", len(page.Nodes), page.TotalCount)
	}
}

func TestFindAllSorting(t *testing.T) {
	repo, all := newTestRepo(t)
	ctx := context.Background()

	for _, dir := range []types.SortDirection{types.SortAsc, types.SortDesc} {
		t.Run(string(dir), func(t *testing.T) {
			// Pages of 2, so the ties (3 Jansens, 2 Bakkers) span pages.
			var got []*models.User
			for offset := 0; offset < len(all); offset += 2 {
				page, err := repo.FindAll(ctx, types.UserListParams{
					Params:  paging.Params{Offset: offset, Limit: 2},
					OrderBy: types.UserOrder{Field: types.UserSortLastName, Direction: dir},
				})
				if err != nil {
					t.Fatal(err)
				}
				got = append(got, page.Nodes...)
			}
			if len(got) != len(all) {
				t.Fatalf("%d users across pages, want %d", len(got), len(all))
			}
			for i := 1; i < len(got); i++ {
				prev, cur := got[i-1], got[i]
				a, b := strings.ToLower(prev.LastName), strings.ToLower(cur.LastName)
				if dir == types.SortDesc {
					a, b = b, a
				}
				if a > b {
					t.Errorf("%q before %q", prev.LastName, cur.LastName)
				}
				if prev.LastName == cur.LastName && prev.ID > cur.ID {
					t.Errorf("tie on %q not broken by ascending id", cur.LastName)
				}
			}
		})
	}

	_, err := repo.FindAll(ctx, types.UserListParams{
		Params:  paging.Params{Limit: 10},
		OrderBy: types.UserOrder{Field: "password; DROP TABLE users", Direction: types.SortAsc},
	})
	if err == nil {
		t.Error("unknown sort field: want an error")
	}
}

func TestFindAllFiltering(t *testing.T) {
	repo, _ := newTestRepo(t)

	tests := []struct {
		name   string
		filter types.UserFilter
		want   int
	}{
		{"no filter", types.UserFilter{}, 15},
		{"last name, any case", types.UserFilter{LastName: "JANSEN"}, 3},
		{"first and last name", types.UserFilter{FirstName: "li", LastName: "jans"}, 1},
		{"search full name", types.UserFilter{Search: "tess de groot"}, 1},
		{"search matches email", types.UserFilter{Search: "vandenberg"}, 1},
		{"search matches phone", types.UserFilter{Search: "+316"}, 5},
		{"underscore is literal", types.UserFilter{Email: "_"}, 1},
		{"percent is literal", types.UserFilter{Search: "%"}, 0},
		{"apostrophe", types.UserFilter{LastName: "o'brien"}, 1},
		{"accent-insensitive collation", types.UserFilter{FirstName: "zoe"}, 1},
		{"no match", types.UserFilter{Email: "nobody"}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, err := repo.FindAll(context.Background(), types.UserListParams{Params: paging.Params{Limit: 50}, Filter: tt.filter})
			if err != nil {
				t.Fatal(err)
			}
			if page.TotalCount != int64(tt.want) || len(page.Nodes) != tt.want {
				t.Errorf("total %d, rows %d, want %d: %v", page.TotalCount, len(page.Nodes), tt.want, emails(page.Nodes))
			}
		})
	}
}

func TestFindByIDsSkipsUnknown(t *testing.T) {
	repo, all := newTestRepo(t)

	users, err := repo.FindByIDs(context.Background(), []string{all[3].ID, "3K4C4mYaiNGD26bPfleDGK7LsNL", all[1].ID})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := emails(users), emails(sortedByID([]*models.User{all[1], all[3]})); !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	users, err = repo.FindByIDs(context.Background(), nil)
	if err != nil || len(users) != 0 {
		t.Errorf("no ids: %v, %v", users, err)
	}
}

func TestErrorsAreDomainErrors(t *testing.T) {
	repo, all := newTestRepo(t)
	ctx := context.Background()
	missing := &models.User{Model: database.Model{ID: "3K4C4mYaiNGD26bPfleDGK7LsNL"}, FirstName: "No", LastName: "One", Email: "no.one@example.com"}

	if _, err := repo.FindByID(ctx, missing.ID); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("FindByID: %v, want ErrUserNotFound", err)
	}
	if _, err := repo.FindByEmail(ctx, missing.Email); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("FindByEmail: %v, want ErrUserNotFound", err)
	}
	if err := repo.Update(ctx, missing); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Update: %v, want ErrUserNotFound", err)
	}
	if err := repo.Delete(ctx, missing.ID); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("Delete: %v, want ErrUserNotFound", err)
	}

	// The unique index catches duplicates the service's check can't see, such
	// as a different case or a soft-deleted user.
	dup := &models.User{FirstName: "Emma", LastName: "Copy", Email: strings.ToUpper(all[0].Email)}
	if err := repo.Create(ctx, dup); !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("Create with a taken email: %v, want ErrEmailTaken", err)
	}
	if err := repo.Delete(ctx, all[1].ID); err != nil {
		t.Fatal(err)
	}
	reuse := &models.User{FirstName: "Liam", LastName: "Again", Email: all[1].Email}
	if err := repo.Create(ctx, reuse); !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("Create with a deleted user's email: %v, want ErrEmailTaken", err)
	}
}

func TestUpdateKeepsCreatedAt(t *testing.T) {
	repo, all := newTestRepo(t)
	ctx := context.Background()
	original := all[0]

	changed := *original
	changed.LastName = "Changed"
	changed.Phone = nil
	changed.CreatedAt = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC) // must be ignored
	if err := repo.Update(ctx, &changed); err != nil {
		t.Fatal(err)
	}

	got, err := repo.FindByID(ctx, original.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.LastName != "Changed" || got.Phone != nil {
		t.Errorf("fields not updated: %+v", got)
	}
	if !got.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("created_at = %v, want unchanged %v", got.CreatedAt, original.CreatedAt)
	}
}

func TestDeletedUsersAreHidden(t *testing.T) {
	repo, all := newTestRepo(t)
	ctx := context.Background()

	if err := repo.Delete(ctx, all[0].ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindByID(ctx, all[0].ID); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("FindByID of a deleted user: %v", err)
	}
	page, err := repo.FindAll(ctx, types.UserListParams{Params: paging.Params{Limit: 50}})
	if err != nil {
		t.Fatal(err)
	}
	if page.TotalCount != int64(len(all)-1) || slices.Contains(emails(page.Nodes), all[0].Email) {
		t.Errorf("deleted user still listed (total %d)", page.TotalCount)
	}
	if err := repo.Delete(ctx, all[0].ID); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("deleting twice: %v, want ErrUserNotFound", err)
	}
}
