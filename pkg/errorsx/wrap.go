package errorsx

import "errors"

// Is reports whether any error in err's chain matches target.
func Is(err, target error) bool { return errors.Is(err, target) }

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true.
func As(err error, target interface{}) bool { return errors.As(err, target) }

// Unwrap returns the result of calling the Unwrap method on err, if err's
// type contains an Unwrap method returning error. Otherwise, Unwrap returns nil.
func Unwrap(err error) error { return errors.Unwrap(err) }
