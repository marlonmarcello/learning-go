package validator

import (
	"slices"
	"strings"
	"unicode/utf8"
)

/*
		lots of examples on validation:
	  https://www.alexedwards.net/blog/validation-snippets-for-go
*/
type Validator struct {
	FormErrors map[string]string
}

func (v *Validator) Valid() bool {
	return len(v.FormErrors) == 0
}

func (v *Validator) AddFormError(key, message string) {
	if v.FormErrors == nil {
		v.FormErrors = make(map[string]string)
	}

	if _, exists := v.FormErrors[key]; !exists {
		v.FormErrors[key] = message
	}
}

func (v *Validator) CheckField(ok bool, key, message string) {
	if !ok {
		v.AddFormError(key, message)
	}
}

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func MaxChars(value string, n int) bool {
	return utf8.RuneCountInString(value) <= n
}

// to learn more about generics
// https://www.youtube.com/watch?v=Pa_e9EeCdy8
// also a tutorial section on the main docs
// https://go.dev/doc/tutorial/generics
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}
