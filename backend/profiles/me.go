package profiles

import (
	"context"
)

// Me returns the profile of the currently authenticated caller, or nil if
// the request is not authenticated.
//
//encore:api public method=GET path=/auth/me
func Me(ctx context.Context) (*Profile, error) {
	return Get(ctx)
}
