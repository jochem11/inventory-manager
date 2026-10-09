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
	"github.com/jochem11/inventory-manager/shared/errs"
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
// tags. Invalid input is returned as an *errs.ValidationError.
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
	root := reflect.TypeOf(v).Elem()
	fields := make(map[string]string, len(fieldErrs))
	for _, fe := range fieldErrs {
		fields[jsonPath(root, fe)] = fe.Translate(trans)
	}
	return &errs.ValidationError{Fields: fields}
}

// jsonPath is the field's path in JSON names, e.g. "home.city", as clients
// send it. Embedded structs (like paging.Params) are left out of the path:
// their fields are promoted, so in JSON they sit on the outer struct.
func jsonPath(root reflect.Type, fe validator.FieldError) string {
	// Both namespaces start with the type name: "Type.field.sub".
	names := strings.Split(fe.Namespace(), ".")[1:]
	goNames := strings.Split(fe.StructNamespace(), ".")[1:]
	t := root
	path := make([]string, 0, len(names))
	for i, goName := range goNames {
		for t.Kind() == reflect.Pointer || t.Kind() == reflect.Slice || t.Kind() == reflect.Array || t.Kind() == reflect.Map {
			t = t.Elem()
		}
		field, ok := reflect.StructField{}, false
		if t.Kind() == reflect.Struct {
			name, _, _ := strings.Cut(goName, "[") // "Items[0]" → "Items"
			field, ok = t.FieldByName(name)
		}
		if ok {
			t = field.Type
			if field.Anonymous {
				continue
			}
		}
		path = append(path, names[i])
	}
	return strings.Join(path, ".")
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
