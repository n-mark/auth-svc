package handler

import (
	"net/http"

	"minimal-service/pkg/response"
)

// HealthHandler responds with {"status": "OK"}.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "OK"})
}
