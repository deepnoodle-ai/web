package errors

import (
	"errors"
	"fmt"
)

// New wraps errors.New
func New(message string) error {
	return errors.New(message)
}

// As wraps errors.As
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Is wraps errors.Is
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// Unwrap wraps errors.Unwrap
func Unwrap(err error) error {
	return errors.Unwrap(err)
}

// BadRequest represents a 400 error.
type BadRequest struct {
	Message string `json:"message"`
}

func (b *BadRequest) Error() string {
	return b.Message
}

func NewBadRequest(message string, args ...any) *BadRequest {
	return &BadRequest{Message: fmt.Sprintf(message, args...)}
}

// NotFound represents a 404 error.
type NotFound struct {
	Message string `json:"message"`
}

func (n *NotFound) Error() string {
	return n.Message
}

func NewNotFound(message string, args ...any) *NotFound {
	return &NotFound{Message: fmt.Sprintf(message, args...)}
}

// Unauthorized represents a 401 error.
type Unauthorized struct {
	Message string `json:"message"`
}

func (u *Unauthorized) Error() string {
	return u.Message
}

func NewUnauthorized(message string, args ...any) *Unauthorized {
	return &Unauthorized{Message: fmt.Sprintf(message, args...)}
}

// Forbidden represents a 403 error.
type Forbidden struct {
	Message string `json:"message"`
}

func (f *Forbidden) Error() string {
	return f.Message
}

func NewForbidden(message string, args ...any) *Forbidden {
	return &Forbidden{Message: fmt.Sprintf(message, args...)}
}

// InternalServerError represents a 500 error.
type InternalServerError struct {
	Message string `json:"message"`
}

func (i *InternalServerError) Error() string {
	return i.Message
}

func NewInternalServerError(message string, args ...any) *InternalServerError {
	return &InternalServerError{Message: fmt.Sprintf(message, args...)}
}

func IsNotFound(err error) bool {
	_, ok := err.(*NotFound)
	return ok
}

func IsBadRequest(err error) bool {
	_, ok := err.(*BadRequest)
	return ok
}

func IsUnauthorized(err error) bool {
	_, ok := err.(*Unauthorized)
	return ok
}

func IsForbidden(err error) bool {
	_, ok := err.(*Forbidden)
	return ok
}

func IsInternalServerError(err error) bool {
	_, ok := err.(*InternalServerError)
	return ok
}

func IsRequestError(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*RequestError)
	return ok
}
