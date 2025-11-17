package types

import (
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

// Error represents a standardized error in the QuantaID system.
// It provides a consistent structure for errors across different layers (e.g., HTTP, gRPC)
// and includes a machine-readable code, a human-readable message, and optional details.
type Error struct {
	// Code is a machine-readable string identifying the error type (e.g., "invalid_credentials").
	Code string `json:"code"`
	// Message is a human-readable description of the error.
	Message string `json:"message"`
	// Details provides additional key-value information about the error.
	Details map[string]string `json:"details,omitempty"`
	// HttpStatus is the corresponding HTTP status code for this error.
	HttpStatus int `json:"-"`
	// GrpcStatus is the corresponding gRPC status code for this error.
	GrpcStatus codes.Code `json:"-"`
	// cause is the underlying error that triggered this error, for internal debugging.
	cause error
}

// Error returns the string representation of the error, satisfying the standard error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("code: %s, message: %s", e.Code, e.Message)
}

// Unwrap returns the wrapped cause of the error, allowing for error chaining.
func (e *Error) Unwrap() error {
	return e.cause
}

// WithCause wraps an existing error, allowing for the preservation of the original error context.
// This is useful for chaining errors without losing the root cause.
func (e *Error) WithCause(cause error) *Error {
	e.cause = cause
	return e
}

// WithDetails adds contextual details to the error, providing more specific information
// about the error condition.
func (e *Error) WithDetails(details map[string]string) *Error {
	e.Details = details
	return e
}

// ToGRPCStatus converts the application-specific error to a gRPC status,
// which can be sent over a gRPC connection.
func (e *Error) ToGRPCStatus() *status.Status {
	return status.New(e.GrpcStatus, e.Message)
}

// NewError creates a new standardized error.
// This function is used to define the standard error types used throughout the application.
func NewError(code, message string, httpStatus int, grpcStatus codes.Code) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HttpStatus: httpStatus,
		GrpcStatus: grpcStatus,
	}
}

// Pre-defined error types for common scenarios.
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
	ErrNotImplemented        = NewError("not_implemented", "This feature is not implemented", http.StatusNotImplemented, codes.Unimplemented)
	ErrInvalidRequest        = NewError("invalid_request", "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed.", http.StatusBadRequest, codes.InvalidArgument)
	ErrInvalidClient         = NewError("invalid_client", "Client authentication failed (e.g., unknown client, no client authentication included, or unsupported authentication method).", http.StatusUnauthorized, codes.Unauthenticated)
	ErrInvalidGrant          = NewError("invalid_grant", "The provided authorization grant (e.g., authorization code, resource owner credentials) or refresh token is invalid, expired, revoked, does not match the redirection URI used in the authorization request, or was issued to another client.", http.StatusBadRequest, codes.InvalidArgument)
	ErrUnsupportedGrantType  = NewError("unsupported_grant_type", "The authorization grant type is not supported by the authorization server.", http.StatusBadRequest, codes.InvalidArgument)
	ErrUserNotFound          = NewError("user_not_found", "The user was not found.", http.StatusNotFound, codes.NotFound)
	ErrSessionExpired        = NewError("session_expired", "The user session has expired.", http.StatusUnauthorized, codes.Unauthenticated)
	ErrDeviceMismatch        = NewError("device_mismatch", "The device fingerprint does not match the session.", http.StatusUnauthorized, codes.Unauthenticated)
	ErrMaxSessionsExceeded   = NewError("max_sessions_exceeded", "The maximum number of concurrent sessions has been exceeded.", http.StatusForbidden, codes.PermissionDenied)
	ErrAuthorizationPending  = NewError("authorization_pending", "The authorization request is still pending as the end-user has not yet completed the user interaction steps.", http.StatusBadRequest, codes.Unavailable)
	ErrAccessDenied          = NewError("access_denied", "The resource owner or authorization server denied the request.", http.StatusForbidden, codes.PermissionDenied)
	ErrExpiredToken          = NewError("expired_token", "The token has expired.", http.StatusBadRequest, codes.Unauthenticated)
)
