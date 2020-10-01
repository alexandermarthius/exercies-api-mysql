package utils

import (
	"encoding/json"
	"net/http"
)

func ResponseJSON(w http.ResponseWriter, body interface{}, status int) {
	responseBody, err := json.Marshal(body)

	if err != nil {
		http.Error(w, "encoding json error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(responseBody))

}
