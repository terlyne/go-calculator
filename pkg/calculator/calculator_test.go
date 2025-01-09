package calculator


import (
	"testing"
	"errors"
)

func TestCalc(t *testing.T) {
	tests := []struct {
		expression string
		expected   float64
		err        error
	}{
		{"3 + 5", 8, nil},
		{"10 + 2 * 6", 22, nil},
		{"100 * 2 + 12", 212, nil},
		{"100 * (2 + 12)", 1400, nil},
		{"100 * (2 + 12) / 14", 100, nil},
		{"3 + 5 * (2 - 4) / 2", 2, nil},
		{"3 + 5 * (2 - 4) / 0", 0, errors.New("Деление на ноль")},
		{"3 + 5 * (2 - 4", 0, errors.New("Несовпадение скобок")},
		{"3 + 5 * (2 - 4))", 0, errors.New("Несовпадение скобок")},
		{"3 + 5 * (2 - 4) / 2 +", 0, errors.New("Ошибка вычисления: неверное количество элементов на стеке")},
		{"3 + 5 * (2 - 4) / a", 0, errors.New("Недопустимый символ в выражении")},
	}

	for _, test := range tests {
		result, err := Calc(test.expression)
		if err != nil && err.Error() != test.err.Error() {
			t.Errorf("Calc(%q) returned error: %v, expected: %v", test.expression, err, test.err)
		}
		if result != test.expected {
			t.Errorf("Calc(%q) = %v, expected %v", test.expression, result, test.expected)
		}
	}
}