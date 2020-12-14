package plugin

import "fmt"

type Error struct {
	err     error
	message string
}

func (l *Error) Error() string {
	return l.err.Error()
}

func (l Error) Message() string {
	return l.message
}

func Err(err error, format string, args ...interface{}) *Error {
	return &Error{
		err:     err,
		message: fmt.Sprintf(format, args...),
	}
}

func InternalErr(err error) *Error {
	return Err(err, "Failed to start session")
}
