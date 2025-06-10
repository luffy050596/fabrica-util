// Package errors provides error handling utilities and wrappers around github.com/pkg/errors
package errors

import (
	"errors"

	pkgerrors "github.com/pkg/errors"
)

// New returns a new error with the given message
// It's a wrapper around github.com/pkg/errors.New
func New(message string) error {
	return pkgerrors.New(message)
}

// Errorf formats according to a format specifier and returns the string as an error
// It's a wrapper around github.com/pkg/errors.Errorf
func Errorf(format string, args ...any) error {
	return pkgerrors.Errorf(format, args...)
}

// Wrap wraps an error with a message
// It's a wrapper around github.com/pkg/errors.Wrap
func Wrap(err error, message string) error {
	return pkgerrors.Wrap(err, message)
}

// Wrapf wraps an error with a formatted message
// It's a wrapper around github.com/pkg/errors.Wrapf
func Wrapf(err error, format string, args ...any) error {
	return pkgerrors.Wrapf(err, format, args...)
}

// WithMessage returns an error that wraps the given error with the given message
// It's a wrapper around github.com/pkg/errors.WithMessage
func WithMessage(err error, message string) error {
	return pkgerrors.WithMessage(err, message)
}

// WithMessagef returns an error that wraps the given error with the given message
// It's a wrapper around github.com/pkg/errors.WithMessagef
func WithMessagef(err error, format string, args ...any) error {
	return pkgerrors.WithMessagef(err, format, args...)
}

// Join returns an error that wraps the given errors
// It's a wrapper around errors.Join
func Join(errs ...error) error {
	return errors.Join(errs...)
}

// JoinUnsimilar returns an error that wraps the given errors,
// but only if the errors are not the same.
// It's a wrapper around errors.Join
func JoinUnsimilar(errs ...error) error {
	err := errs[0]

	for _, e := range errs[1:] {
		if e == nil {
			continue
		}

		if errors.Is(err, e) {
			continue
		}

		err = errors.Join(err, e)
	}

	return err
}

// Is reports whether any error in err's tree matches target
// It's a wrapper around github.com/pkg/errors.Is
func Is(err, target error) bool {
	return pkgerrors.Is(err, target)
}

// As finds the first error in err's tree that matches target, and if one is found, sets
// target to that error value and returns true. Otherwise, it returns false.
// It's a wrapper around github.com/pkg/errors.As
func As(err error, target any) bool {
	return pkgerrors.As(err, target)
}

// Unwrap returns the underlying error of err, if there is one.
// It's a wrapper around github.com/pkg/errors.Unwrap
func Unwrap(err error) error {
	return pkgerrors.Unwrap(err)
}
