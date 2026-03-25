package middleware

import (
	"context"
	"net/http"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("id")
		var userId string

		if err == nil {
			userId = cookie.Value
		}

		ctx := context.WithValue(r.Context(), "id", userId)
		next(w, r.WithContext(ctx))
	}
}
