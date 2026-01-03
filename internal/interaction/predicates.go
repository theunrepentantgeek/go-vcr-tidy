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
