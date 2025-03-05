package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Task struct {
	ID        string  `json:"id"`
	Arg1      float64 `json:"arg1"`
	Arg2      float64 `json:"arg2"`
	Operation string  `json:"operation"`
}

func StartAgent() {
	for {
		resp, err := http.Get("http://localhost/internal/task")
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			var task Task
			if err := json.NewDecoder(resp.Body).Decode(&task); err == nil {
				result := performOperation(task)
				http.Post("http://localhost/internal/task", "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"id": "%v", "result": %v}`, task.ID, result))))
			}
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

func performOperation(task Task) float64 {
	switch task.Operation {
	case "+":
		return task.Arg1 + task.Arg2
	case "-":
		return task.Arg1 - task.Arg2
	case "*":
		return task.Arg1 * task.Arg2
	case "/":
		return task.Arg1 / task.Arg2
	default:
		return 0
	}
}
