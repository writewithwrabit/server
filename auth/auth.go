package auth

import (
	"context"
	"log"
	"net/http"
	"strings"

	"firebase.google.com/go/auth"
)

// A private key for context that only this package can access. This is important
// to prevent collisions between different context uses
var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

// Middleware decodes the share session cookie and packs the session into context
func Middleware(client *auth.Client) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token *auth.Token
			t := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(t) == 2 || t[0] == "Bearer" {
				var err error
				token, err = client.VerifyIDToken(context.Background(), t[1])
				if err != nil {
					http.Error(w, "Invalid token", http.StatusForbidden)
					return
				}

				log.Printf("Verified ID token", token)
			}

			// put it in context
			ctx := context.WithValue(r.Context(), userCtxKey, token)

			// and call the next with our new context
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}

// ForContext finds the user from the context. REQUIRES Middleware to have run.
func ForContext(ctx context.Context) *auth.Token {
	raw, _ := ctx.Value(userCtxKey).(*auth.Token)
	return raw
}
