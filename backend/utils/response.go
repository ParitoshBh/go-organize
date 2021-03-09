package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// SendJSONResponse prepare and send json response
func SendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// GetReponseRedirect builds redirect slug along with query params
func GetReponseRedirect(currentPath string) string {
	if currentPath == "" {
		return "/"
	}

	// Strip trailing "/" if any
	if currentPath[len(currentPath)-1:] == "/" {
		currentPath = currentPath[0 : len(currentPath)-1]
	}

	return fmt.Sprintf("/?path=%s", currentPath)
}
