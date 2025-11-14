package utils

import (
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
