package middleware

import (
	"context"
	"github.com/argad/url-shortener/internal/auth"
	"net/http"
	"time"
)

type contextKey string

// UserIDKey is the key used to store the user ID in the request context.
const UserIDKey contextKey = "userID"

// AuthMiddleware is a middleware that handles user authentication.
// It checks for a valid JWT token in the request cookie and creates a new user if one doesn't exist.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string

		cookie, err := r.Cookie(auth.CookieName)
		if err == nil {
			if id, verifyErr := auth.VerifyJWTToken(cookie.Value); verifyErr == nil {
				userID = id
			}
		}

		if userID == "" {
			newUserID, genErr := auth.GenerateUserID()
			if genErr != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			userID = newUserID

			jwtToken, tokenErr := auth.CreateJWTToken(userID)
			if tokenErr != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{
				Name:     auth.CookieName,
				Value:    jwtToken,
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(24 * time.Hour * 365),
			})
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID retrieves the user ID from the request context.
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// RequireAuth is a middleware that requires the user to be authenticated.
// It checks for a valid user ID in the request context and returns an unauthorized error if one doesn't exist.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok || userID == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
