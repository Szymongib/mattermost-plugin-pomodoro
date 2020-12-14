package plugin

import "fmt"

type PluginError struct {
	err     error
	message string
}

func (l *PluginError) Error() string {
	return l.err.Error()
}

func (l PluginError) Message() string {
	return l.message
}

func Err(err error, format string, args ...interface{}) *PluginError {
	return &PluginError{
		err:     err,
		message: fmt.Sprintf(format, args...),
	}
}

func InternalErr(err error) *PluginError {
	return Err(err, "Failed to start session")
}
