package orchestrator

import (
	"testing"
)

func TestAddExpression(t *testing.T) {
	AddExpression("test_id")
	expr, exists := GetExpressionByID("test_id")
	if !exists || expr.Status != "pending" {
		t.Error(expr.Status)
	}
}
