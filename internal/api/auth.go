package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func (h *Handlers) handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid json"}`, 400)
			return
		}
		if req.Password != h.JWTSecret {
			http.Error(w, `{"error":"unauthorized"}`, 401)
			return
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(24 * time.Hour).Unix(),
		})
		tokenStr, _ := token.SignedString([]byte(h.JWTSecret))
		json.NewEncoder(w).Encode(map[string]string{"token": tokenStr})
	}
}

func (h *Handlers) authMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := r.Header.Get("Authorization")
			if len(tokenStr) < 8 || tokenStr[:7] != "Bearer " {
				http.Error(w, `{"error":"missing token"}`, 401)
				return
			}
			token, err := jwt.Parse(tokenStr[7:], func(t *jwt.Token) (interface{}, error) {
				return []byte(h.JWTSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, `{"error":"invalid token"}`, 401)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
