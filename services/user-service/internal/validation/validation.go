// Package validation cleans up and validates input structs using their tags:
// `mod` (go-playground/mold) to normalize, then `validate`
// (go-playground/validator) to check. Field names in errors come from the
// `json` tags, so they match what clients send.
package validation

import (
	"context"
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/mold/v4/modifiers"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	entranslations "github.com/go-playground/validator/v10/translations/en"
	"github.com/jochem11/inventory-manager/services/user-service/internal/domain"
)

// Both cache struct metadata, so they are shared rather than created per call.
var (
	conform  = modifiers.New()
	validate = validator.New(validator.WithRequiredStructEnabled())
	trans    ut.Translator
)

func init() {
	validate.RegisterTagNameFunc(func(f reflect.StructField) string {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if name == "-" {
			return ""
		}
		return name
	})

	english := en.New()
	trans, _ = ut.New(english, english).GetTranslator("en")
	if err := entranslations.RegisterDefaultTranslations(validate, trans); err != nil {
		panic(err)
	}
	// The default translations don't cover http_url.
	if err := validate.RegisterTranslation("http_url", trans,
		func(ut ut.Translator) error { return ut.Add("http_url", "{0} must be an http(s) URL", false) },
		func(ut ut.Translator, fe validator.FieldError) string { t, _ := ut.T("http_url", fe.Field()); return t },
	); err != nil {
		panic(err)
	}
}

// Clean normalizes v (a pointer to a struct) by its `mod` tags, turns optional
// strings that end up empty into nil, then validates it by its `validate`
// tags. Invalid input is returned as a *domain.ValidationError.
func Clean(ctx context.Context, v any) error {
	if err := conform.Struct(ctx, v); err != nil {
		return err
	}
	nilEmptyStrings(reflect.ValueOf(v).Elem())

	err := validate.Struct(v)
	var fieldErrs validator.ValidationErrors
	if !errors.As(err, &fieldErrs) {
		return err
	}
	fields := make(map[string]string, len(fieldErrs))
	for _, fe := range fieldErrs {
		// Namespace is "Type.field.sub"; drop the type to key by the JSON path.
		_, path, _ := strings.Cut(fe.Namespace(), ".")
		fields[path] = fe.Translate(trans)
	}
	return &domain.ValidationError{Fields: fields}
}

// nilEmptyStrings sets *string fields that point to "" to nil, so an optional
// field left blank is stored as "not set". mold can't do this: its modifiers
// only see the value behind a pointer, not the pointer itself.
func nilEmptyStrings(v reflect.Value) {
	for i := range v.NumField() {
		f := v.Field(i)
		switch {
		case !f.CanSet():
		case f.Kind() == reflect.Pointer && f.Type().Elem().Kind() == reflect.String:
			if !f.IsNil() && f.Elem().String() == "" {
				f.SetZero()
			}
		case f.Kind() == reflect.Struct:
			nilEmptyStrings(f)
		case f.Kind() == reflect.Pointer && f.Type().Elem().Kind() == reflect.Struct && !f.IsNil():
			nilEmptyStrings(f.Elem())
		}
	}
}
