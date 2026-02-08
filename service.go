package main

// Shared errors used by handlers and answer service.
var (
	ErrQuestionNotFound = &Error{Message: "question not found"}
	ErrDuplicateAnswer  = &Error{Message: "duplicate answer"}
)

// Error is a simple error type for quiz errors.
type Error struct {
	Message string
}

func (e *Error) Error() string {
	return e.Message
}
