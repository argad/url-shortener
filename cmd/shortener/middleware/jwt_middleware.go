package middleware

import (
	"context"
	"github.com/argad/url-shortener/cmd/shortener/auth"
	"net/http"
	"time"
)

type contextKey string

const UserIDKey contextKey = "userID"

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

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

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
