package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/nabsk911/chronify/internal/auth"
	"github.com/nabsk911/chronify/internal/utils"
)

func Authentication(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Authorization header required!"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenString == authHeader {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Bearer token required!"})
			return
		}

		claims, err := auth.ValidateToken(tokenString)

		if err != nil {
			utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "Invalid token!"})
			return
		}

		ctx := context.WithValue(r.Context(), "userID", claims.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
