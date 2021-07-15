package common

type ErrorCode int

const (
	GENERAL_ERR ErrorCode = iota
	NOT_FOUND
)

type Error struct {
	msg  string
	code ErrorCode
}

func NewError(message string) *Error {
	return &Error{message, GENERAL_ERR}
}

func NewErrorWithCode(code ErrorCode, message string) *Error {
	return &Error{message, code}
}

func (e *Error) Error() string {
	return e.msg
}

func (e *Error) Code() ErrorCode {
	return e.code
}
