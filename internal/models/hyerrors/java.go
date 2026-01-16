package hyerrors

import "errors"

var (
	ErrJavaNotFound = errors.New("java not found")
	ErrJavaBroken   = errors.New("java is broken")
)
