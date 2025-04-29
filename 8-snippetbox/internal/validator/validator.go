package validator

import (
	"regexp"
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

// Use the regexp.MustCompile() function to parse a regular expression pattern
// for sanity checking the format of an email address. This returns a pointer to
// a 'compiled' regexp.Regexp type, or panics in the event of an error. Parsing
// this pattern once at startup and storing the compiled *regexp.Regexp in a
// variable is more performant than re-parsing the pattern each time we need it.

// The pattern we’re using is the one currently recommended by the W3C and Web Hypertext Application Technology Working Group for validating email addresses. For more information about this pattern, see here. If you’re reading this book in PDF format or on a narrow device, and can’t see the entire line, then here it is broken up into multiple lines:
var EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

func MinChars(value string, n int) bool {
	return utf8.RuneCountInString(value) >= n
}

func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}
