package types

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

// Error represents a standardized error in the QuantaID system.
type Error struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    map[string]string `json:"details,omitempty"`
	HttpStatus int               `json:"-"`
	GrpcStatus codes.Code        `json:"-"`
	cause      error
}

// Error returns the string representation of the error.
func (e *Error) Error() string {
	return fmt.Sprintf("code: %s, message: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped cause of the error.
func (e *Error) Unwrap() error {
	return e.cause
}

// WithCause wraps an existing error.
func (e *Error) WithCause(cause error) *Error {
	e.cause = cause
	return e
}

// WithDetails adds contextual details to the error.
func (e *Error) WithDetails(details map[string]string) *Error {
	e.Details = details
	return e
}

// ToGRPCStatus converts the error to a gRPC status.
func (e *Error) ToGRPCStatus() *status.Status {
	return status.New(e.GrpcStatus, e.Message)
}

// NewError creates a new standardized error.
func NewError(code, message string, httpStatus int, grpcStatus codes.Code) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HttpStatus: httpStatus,
		GrpcStatus: grpcStatus,
	}
}

var (
	ErrInternal              = NewError("internal_error", "An unexpected internal error occurred", http.StatusInternalServerError, codes.Internal)
	ErrNotFound              = NewError("not_found", "The requested resource was not found", http.StatusNotFound, codes.NotFound)
	ErrBadRequest            = NewError("bad_request", "The request is malformed or invalid", http.StatusBadRequest, codes.InvalidArgument)
	ErrValidation            = NewError("validation_failed", "Input validation failed", http.StatusBadRequest, codes.InvalidArgument)
	ErrUnauthorized          = NewError("unauthorized", "Authentication is required and has failed or has not yet been provided", http.StatusUnauthorized, codes.Unauthenticated)
	ErrForbidden             = NewError("forbidden", "You do not have permission to access this resource", http.StatusForbidden, codes.PermissionDenied)
	ErrConflict              = NewError("conflict", "The resource already exists", http.StatusConflict, codes.AlreadyExists)
	ErrTooManyRequests       = NewError("too_many_requests", "You have exceeded the rate limit", http.StatusTooManyRequests, codes.ResourceExhausted)
	ErrServiceUnavailable    = NewError("service_unavailable", "The service is temporarily unavailable", http.StatusServiceUnavailable, codes.Unavailable)
	ErrInvalidCredentials    = NewError("invalid_credentials", "Invalid username or password", http.StatusUnauthorized, codes.Unauthenticated)
	ErrInvalidToken          = NewError("invalid_token", "The provided token is invalid or expired", http.StatusUnauthorized, codes.Unauthenticated)
	ErrTokenExpired          = NewError("token_expired", "The provided token has expired", http.StatusUnauthorized, codes.Unauthenticated)
	ErrMfaRequired           = NewError("mfa_required", "Multi-factor authentication is required", http.StatusUnauthorized, codes.Unauthenticated)
	ErrMfaChallengeInvalid   = NewError("mfa_challenge_invalid", "The MFA challenge is invalid or has expired", http.StatusBadRequest, codes.InvalidArgument)
	ErrUserLocked            = NewError("user_locked", "The user account is locked", http.StatusForbidden, codes.PermissionDenied)
	ErrUserDisabled          = NewError("user_disabled", "The user account is disabled", http.StatusForbidden, codes.PermissionDenied)
	ErrPluginLoadFailed      = NewError("plugin_load_failed", "Failed to load plugin", http.StatusInternalServerError, codes.Internal)
	ErrPluginNotFound        = NewError("plugin_not_found", "The requested plugin was not found", http.StatusNotFound, codes.NotFound)
	ErrPluginInitFailed      = NewError("plugin_init_failed", "Failed to initialize plugin", http.StatusInternalServerError, codes.Internal)
)

//Personal.AI order the ending
