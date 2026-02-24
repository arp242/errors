//go:build go1.26

package errors

import "errors"

func AsType[E error](err error) (E, bool) { return errors.AsType[E](err) }
