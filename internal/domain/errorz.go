package domain

import "fmt"

type ErrorCode string

const (
	ErrCodeTeamExists   ErrorCode = "TEAM_EXISTS"
	ErrCodePRExists     ErrorCode = "PR_EXISTS"
	ErrCodePRMerged     ErrorCode = "PR_MERGED"
	ErrCodeNotAssigned  ErrorCode = "NOT_ASSIGNED"
	ErrCodeNoCandidate  ErrorCode = "NO_CANDIDATE"
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeInvalidInput ErrorCode = "INVALID_INPUT"
)

type DomainError struct {
	Code    ErrorCode
	Message string
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}
