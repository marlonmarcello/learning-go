package main

// This defines a new distinct type named contextKey with the underlying type as string
type contextKey string

// contextKey() here isn't a function it's a type conversion
// Go allows conversions between types if their underlying types are compatible (or in specific other cases). Here, we are converting a value that looks like a string into the specific contextKey type.
// When making comparisons, go with check type and values so isAuthenticatedContextKey == "isAuthenticated" will be false
const isAuthenticatedContextKey = contextKey("isAuthenticated")

const authenticatedUserIDSessionKey = "authenticatedUserId"
