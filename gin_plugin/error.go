package gin_plugin

import (
	"errors"
	"fmt"
)

type EZErr struct {
	err     error
	code    int
	message string
}

func (e *EZErr) Error() string {
	return fmt.Sprintf("EZErr [%d] %s", e.code, e.message)
}

func (e *EZErr) WithDetail(message string) *EZErr {
	return &EZErr{
		err:     e.err,
		code:    e.code,
		message: fmt.Sprintf("%s(%s)", e.message, message),
	}
}

func (e *EZErr) WithErr(err error) *EZErr {
	return &EZErr{
		err:     e.err,
		code:    e.code,
		message: fmt.Sprintf("%s: %s", e.message, err.Error()),
	}
}

func (e *EZErr) Is(err error) bool {
	if dstErr, ok := err.(*EZErr); ok {
		return e.code == dstErr.code
	}

	if v, ok := err.(interface{ Is(error) bool }); ok {
		return v.Is(e)
	}

	return errors.Is(e.err, err)
}

func (e *EZErr) Unwrap() error {
	return e.err
}

func (e *EZErr) Code() int {
	return e.code
}

func Error(code int, message string) *EZErr {
	return &EZErr{
		err:     errors.New(message),
		code:    code,
		message: message,
	}
}
