package middleware

import (
	"363project/controller/service"
	"363project/model"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("id")
		if err != nil {
			newUser, err := service.CreateAnonymousUser()
			if err != nil {
				http.Error(w, "Internal Server Error", 500)
				return
			}

			cookieData := model.USSDCookie{
				UserId: newUser.User_id,
				Step:   0,
			}
			jsonBytes, _ := json.Marshal(&cookieData)
			encoded := base64.StdEncoding.EncodeToString(jsonBytes)
			http.SetCookie(w, &http.Cookie{
				Name:     "ussd_state",
				Value:    encoded,
				Path:     "/",
				HttpOnly: true,
				MaxAge:   300,
			})

			ctx := context.WithValue(r.Context(), "ussd", cookieData)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		var cookieData model.USSDCookie

		decoded, err := base64.StdEncoding.DecodeString(cookie.Value)
		if err != nil {
			http.Error(w, "Cookie tidak valid", 400)
			return
		}
		err = json.Unmarshal(decoded, &cookieData)
		if err != nil {
			http.Error(w, "Cookie tidak valid", 400)
			return
		}
		ctx := context.WithValue(r.Context(), "ussd", cookieData)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
