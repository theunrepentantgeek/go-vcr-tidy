package interaction

// Predicates to make common checks on interactions.

import "slices"

// HasMethod checks if the interaction uses the specified HTTP method.
func HasMethod(
	i Interface,
	method string,
) bool {
	return i.Request().Method() == method
}

// HasAnyMethod checks if the interaction uses any of the specified HTTP methods.
func HasAnyMethod(
	i Interface,
	methods ...string,
) bool {
	return slices.Contains(methods, i.Request().Method())
}

// WasSuccessful checks if the interaction's response status code indicates success (2xx).
func WasSuccessful(
	i Interface,
) bool {
	statusCode := i.Response().StatusCode()

	return statusCode >= 200 && statusCode < 300
}
