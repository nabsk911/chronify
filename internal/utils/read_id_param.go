package utils

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
)

func ReadIDParam(r *http.Request, paramName string) (pgtype.UUID, error) {
	idParam := r.PathValue(paramName)
	var id pgtype.UUID
	err := id.Scan(idParam)
	if err != nil {
		return id, err
	}
	return id, nil
}

func ReadUserIDFromContext(r *http.Request, contextKey string) (pgtype.UUID, error) {
	userIDStr, ok := r.Context().Value(contextKey).(string)
	if !ok {
		return pgtype.UUID{}, fmt.Errorf("user ID not found in context or invalid type")
	}

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		return pgtype.UUID{}, fmt.Errorf("invalid user ID format: %w", err)
	}

	return userID, nil
}

