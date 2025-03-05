package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type MockOrchestrator struct {
	tasks   []Task
	results []struct {
		ID     string  `json:"id"`
		Result float64 `json:"result"`
	}
	mu sync.Mutex
}

func (m *MockOrchestrator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if r.Method == http.MethodGet && r.URL.Path == "/internal/task" {
		if len(m.tasks) > 0 {
			task := m.tasks[0]
			m.tasks = m.tasks[1:]
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(task)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
		return
	}

	if r.Method == http.MethodPost && r.URL.Path == "/internal/task" {
		var result struct {
			ID     string  `json:"id"`
			Result float64 `json:"result"`
		}
		if err := json.NewDecoder(r.Body).Decode(&result); err == nil {
			m.results = append(m.results, result)
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
}

func TestAgent(t *testing.T) {
	// создаем тестовый сервер с мок-оркестратором
	mockOrchestrator := &MockOrchestrator{
		tasks: []Task{
			{ID: "1", Arg1: 3, Arg2: 5, Operation: "+"},
			{ID: "2", Arg1: 10, Arg2: 2, Operation: "-"},
		},
	}
	server := httptest.NewServer(mockOrchestrator)
	defer server.Close()

	// изменяем URL агента на адрес тестового сервера
	agentURL := server.URL + "/internal/task"

	// запускаем агента в отдельной горутине
	go func() {
		for {
			resp, err := http.Get(agentURL)
			if err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			if resp.StatusCode == http.StatusOK {
				var task Task
				if err := json.NewDecoder(resp.Body).Decode(&task); err == nil {
					// Выполнить задачу
					result := performOperation(task)
					// Отправить результат
					http.Post(agentURL, "application/json", bytes.NewBuffer([]byte(fmt.Sprintf(`{"id": "%v", "result": %v}`, task.ID, result))))
				}
			} else {
				time.Sleep(1 * time.Second)
			}
		}
	}()

	// даем агенту время для выполнения задач
	time.Sleep(2 * time.Second)

	// проверяем, что результаты были отправлены
	mockOrchestrator.mu.Lock()
	defer mockOrchestrator.mu.Unlock()

	if len(mockOrchestrator.results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(mockOrchestrator.results))
	}

	expectedResults := []struct {
		ID     string
		Result float64
	}{
		{"1", 8}, // 3 + 5
		{"2", 8}, // 10 - 2
	}

	for i, expected := range expectedResults {
		if mockOrchestrator.results[i].ID != expected.ID || mockOrchestrator.results[i].Result != expected.Result {
			t.Errorf("Expected result %v, got %v", expected, mockOrchestrator.results[i])
		}
	}
}
