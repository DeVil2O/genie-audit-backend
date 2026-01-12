package helpers

import "fmt"

type Error struct {
	statusCode  int
	message     string
	internalErr error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s, status code:%d", e.message, e.statusCode)
}

func (e *Error) FullError() string {
	return fmt.Sprintf("%s, status code:%d, internal err: %v", e.message, e.statusCode, e.internalErr)
}

func (e *Error) StatusCode() int {
	return e.statusCode
}

func (e *Error) InternalError() error {
	return e.internalErr
}

func New(message string, statusCode int) *Error {
	return &Error{
		statusCode: statusCode,
		message:    message,
	}
}

func NewWithInternalError(err error, message string, statusCode int) *Error {
	return &Error{
		statusCode:  statusCode,
		message:     message,
		internalErr: err,
	}
}
