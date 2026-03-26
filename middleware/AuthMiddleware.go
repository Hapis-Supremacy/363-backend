package middleware

import (
	"363project/controller/service"
	"363project/model"
	"context"
	"encoding/json"
	"net/http"
)

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("id")
		if err != nil {
			newUser, err := service.CreateAnonymousUser()
			if err != nil {
				http.Error(w, "Internal Server Error", 500)
				return
			}

			cookieData := model.USSDCookie{
				UserId: newUser.Id,
				Step:   0,
			}
			jsonBytes, _ := json.Marshal(&cookieData)
			http.SetCookie(w, &http.Cookie{
				Name:     "ussd_state",
				Value:    string(jsonBytes),
				Path:     "/",
				HttpOnly: true,
				MaxAge:   300,
			})

			ctx := context.WithValue(r.Context(), "ussd", cookieData)
			next(w, r.WithContext(ctx))
			return
		}

		var cookieData model.USSDCookie
		err = json.Unmarshal([]byte(cookie.Value), &cookieData)
		if err != nil {
			http.Error(w, "Cookie tidak valid", 400)
			return
		}
		ctx := context.WithValue(r.Context(), "ussd", cookieData)
		next(w, r.WithContext(ctx))
	}
}
