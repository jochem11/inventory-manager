package validation

import (
	"context"
	"errors"
	"testing"

	"github.com/jochem11/inventory-manager/shared/errs"
	"github.com/jochem11/inventory-manager/shared/paging"
)

type address struct {
	City string  `json:"city" mod:"trim" validate:"required"`
	Note *string `json:"note" mod:"trim"`
}

type form struct {
	Name     string   `json:"name" mod:"trim" validate:"required,max=5"`
	Email    string   `json:"email" mod:"trim,lcase" validate:"omitempty,email"`
	Website  *string  `json:"website" mod:"trim" validate:"omitempty,http_url"`
	Nickname *string  `json:"nickname" mod:"trim"`
	Secret   string   `json:"-" validate:"omitempty,len=3"`
	Home     address  `json:"home"`
	Work     *address `json:"work"`
}

func ptr(s string) *string { return &s }

func TestCleanNormalizes(t *testing.T) {
	f := &form{
		Name:     "  Ada ",
		Email:    " Ada@Example.COM ",
		Nickname: ptr("   "),
		Home:     address{City: " Utrecht ", Note: ptr(" ")},
		Work:     &address{City: "Delft", Note: ptr(" 3rd floor ")},
	}
	if err := Clean(context.Background(), f); err != nil {
		t.Fatal(err)
	}

	if f.Name != "Ada" || f.Email != "ada@example.com" || f.Home.City != "Utrecht" {
		t.Errorf("strings not cleaned: %+v", f)
	}
	if f.Nickname != nil || f.Home.Note != nil {
		t.Error("blank optional strings should become nil, also in nested structs")
	}
	if f.Work.Note == nil || *f.Work.Note != "3rd floor" {
		t.Errorf("non-blank optional string = %v, want trimmed", f.Work.Note)
	}
}

func TestCleanReportsFieldsByJSONPath(t *testing.T) {
	f := &form{
		Name:    "Too long",
		Email:   "nope",
		Website: ptr("ftp://example.com"),
		Secret:  "x",
		Home:    address{City: "  "},
		Work:    &address{},
	}
	err := Clean(context.Background(), f)

	var invalid *errs.ValidationError
	if !errors.As(err, &invalid) {
		t.Fatalf("want an *errs.ValidationError, got %v", err)
	}
	want := map[string]string{
		"name":      "name must be a maximum of 5 characters in length",
		"email":     "email must be a valid email address",
		"website":   "website must be an http(s) URL",
		"Secret":    "Secret must be 3 characters in length",
		"home.city": "city is a required field",
		"work.city": "city is a required field",
	}
	for field, msg := range want {
		if got := invalid.Fields[field]; got != msg {
			t.Errorf("Fields[%q] = %q, want %q", field, got, msg)
		}
	}
	if len(invalid.Fields) != len(want) {
		t.Errorf("got fields %v, want exactly %d", invalid.Fields, len(want))
	}
}

func TestCleanValid(t *testing.T) {
	f := &form{Name: "Ada", Website: ptr("https://example.com"), Home: address{City: "Utrecht"}}
	if err := Clean(context.Background(), f); err != nil {
		t.Fatalf("valid input: %v", err)
	}
}

// Tags on an embedded struct, like paging.Params in a list's parameters,
// apply as if the fields were declared directly.
func TestCleanEmbeddedPaging(t *testing.T) {
	type list struct {
		paging.Params
		Search string `json:"search" mod:"trim"`
	}

	params := list{Search: "  ada "}
	if err := Clean(context.Background(), &params); err != nil {
		t.Fatal(err)
	}
	if params.Limit != 10 || params.Search != "ada" {
		t.Errorf("got limit %d, search %q; want 10 and \"ada\"", params.Limit, params.Search)
	}

	params = list{Params: paging.Params{Offset: -1, Limit: 500}}
	var invalid *errs.ValidationError
	if err := Clean(context.Background(), &params); !errors.As(err, &invalid) ||
		invalid.Fields["offset"] == "" || invalid.Fields["limit"] == "" {
		t.Errorf("want errors for offset and limit, got %v", err)
	}
}
