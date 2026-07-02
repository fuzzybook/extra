package admin

import (
	authsvc "encore.app/auth"
	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
	"encore.dev/middleware"
)

// RequireAdmin ensures the caller is logged in and has the admin role
// before any endpoint in this service runs.
//
//encore:middleware target=all
func RequireAdmin(req middleware.Request, next middleware.Next) middleware.Response {
	if _, ok := auth.UserID(); !ok {
		return middleware.Response{Err: &errs.Error{Code: errs.Unauthenticated, Message: "must be logged in"}}
	}
	data, _ := auth.Data().(*authsvc.AuthData)
	if data == nil || !data.Role.IsAdmin() {
		return middleware.Response{Err: &errs.Error{Code: errs.PermissionDenied, Message: "admin role required"}}
	}
	return next(req)
}
