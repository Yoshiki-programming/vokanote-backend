package responses

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ErrorResponse struct {
	Status  string `json:"status"`
	Message any    `json:"message"`
}

func SendResponse(w http.ResponseWriter, statusCode int, responseBody []byte) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(responseBody)
}

func SendErrorResponse(w http.ResponseWriter, statusCode int, err any) {
	errorResponseJson, err := json.Marshal(&ErrorResponse{
		Status:  fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode)),
		Message: err,
	})
	if err != nil {
		fmt.Print("json can not marshal.")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(errorResponseJson)
}
