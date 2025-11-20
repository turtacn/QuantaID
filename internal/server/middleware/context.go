package middleware

type contextKey string

const (
	UserIDContextKey contextKey = "user_id"
	GroupsContextKey contextKey = "groups"
)
