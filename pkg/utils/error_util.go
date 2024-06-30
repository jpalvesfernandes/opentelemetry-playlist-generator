package utils

import (
	"net/http"
)

func WriteErrorResponse(w http.ResponseWriter, status int, message string) {
	http.Error(w, message, status)
}
