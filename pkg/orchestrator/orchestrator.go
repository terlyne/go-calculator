package orchestrator

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/mux"
	"github.com/terlyne/go-calculator/pkg/calculator"
)

type Expression struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Result *string `json:"result,omitempty"`
}

var (
	expressions = make(map[string]*Expression)
	mu          sync.Mutex
)

func AddExpression(id string) {
	mu.Lock()
	defer mu.Unlock()
	expressions[id] = &Expression{ID: id, Status: "pending"}
}

func GetExpressions() []Expression {
	mu.Lock()
	defer mu.Unlock()
	var result []Expression
	for _, expr := range expressions {
		result = append(result, *expr)
	}
	return result
}

func GetExpressionByID(id string) (*Expression, bool) {
	mu.Lock()
	defer mu.Unlock()
	expr, exists := expressions[id]
	return expr, exists
}

func UpdateExpression(id string, result string) {
	mu.Lock()
	defer mu.Unlock()
	if expr, exists := expressions[id]; exists {
		expr.Status = "completed"
		expr.Result = &result
	}
}

func StartServer() {
	r := mux.NewRouter()

	r.HandleFunc("/api/v1/calculate", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Expression string `json:"expression"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
			return
		}
		id := "expr_" + strconv.Itoa(len(expressions)+1)
		AddExpression(id)

		go func() {
			// логика для вычисления
			result, err := calculateExpression(req.Expression)
			if err != nil {
				UpdateExpression(id, "Error: "+err.Error())
			} else {
				UpdateExpression(id, strconv.FormatFloat(result, 'f', -1, 64))
			}
		}()
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": id})
	}).Methods("POST")

	r.HandleFunc("/api/v1/expressions", func(w http.ResponseWriter, r *http.Request) {
		exprs := GetExpressions()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string][]Expression{"expressions": exprs})
	}).Methods("GET")

	r.HandleFunc("/api/v1/expressions/{id}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		expr, exists := GetExpressionByID(id)
		if !exists {
			http.Error(w, `{"error": "Expression not found"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]Expression{"expression": *expr})
	}).Methods("GET")

	r.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		// Найти первое выражение со статусом "pending"
		for _, expr := range expressions {
			if expr.Status == "pending" {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(expr)
				return
			}
		}

		http.Error(w, `{"error": "No tasks available"}`, http.StatusNotFound)
	}).Methods("GET")

	r.HandleFunc("/internal/task", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     string  `json:"id"`
			Result float64 `json:"result"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "Invalid request body"}`, http.StatusBadRequest)
			return
		}

		UpdateExpression(req.ID, strconv.FormatFloat(req.Result, 'f', -1, 64))

		w.WriteHeader(http.StatusOK)
	}).Methods("POST")

	http.ListenAndServe(":8080", r)
}

func calculateExpression(expression string) (float64, error) {
	return calculator.Calc(expression)
}
