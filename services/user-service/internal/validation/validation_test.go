package validation

import (
	"context"
	"errors"
	"testing"

	"github.com/jochem11/inventory-manager/services/user-service/internal/domain"
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

	var invalid *domain.ValidationError
	if !errors.As(err, &invalid) {
		t.Fatalf("want a *domain.ValidationError, got %v", err)
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
