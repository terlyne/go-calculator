package main

import (
	"calculator/pkg/calculator"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CalculateRequest struct {
	Expression string `json:"expression"`
}

type CalculateResponse struct {
	Result string `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

func calculateHandler(w http.ResponseWriter, r *http.Request) {
	var req CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
		return
	}

	result, err := calculator.Calc(req.Expression)
	w.Header().Set("Content-Type", "application/json")

	if err != nil {
		var statusCode int
		switch err.Error() {
		case "Invalid character in expression":
			statusCode = http.StatusUnprocessableEntity
		default:
			statusCode = http.StatusInternalServerError
		}
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(CalculateResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(CalculateResponse{Result: strconv.FormatFloat(result, 'f', -1, 64)})
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/calculate", calculateHandler).Methods("POST")

	http.ListenAndServe(":8080", r)
}
