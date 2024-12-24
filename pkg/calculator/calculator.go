package calculator

import (
	"errors"
	"strconv"
	"strings"
)

func Calc(expression string) (float64, error) {
	expression = strings.Replace(expression, " ", "", -1)
	tokens, err := tokenize(expression)
	if err != nil {
		return 0, err
	}
	rpn, err := toRPN(tokens)
	if err != nil {
		return 0, err
	}
	result, err := evaluateRPN(rpn)
	if err != nil {
		return 0, err
	}
	return result, nil
}

func tokenize(expression string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	for _, c := range expression {
		switch {
		case c >= '0' && c <= '9' || c == '.':
			current.WriteRune(c)
		case c == '+' || c == '-' || c == '*' || c == '/' || c == '(' || c == ')':
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			tokens = append(tokens, string(c))
		default:
			return nil, errors.New("Недопустимый символ в выражении")
		}
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens, nil
}

func toRPN(tokens []string) ([]string, error) {
	var rpn []string
	var stack []string
	precedence := map[string]int{
		"+": 1, "-": 1, "*": 2, "/": 2,
	}
	for _, token := range tokens {
		switch token {
		case "+", "-", "*", "/":
			for len(stack) > 0 {
				top := stack[len(stack)-1]
				if top == "(" || precedence[top] < precedence[token] {
					break
				}
				rpn = append(rpn, top)
				stack = stack[:len(stack)-1]
			}
			stack = append(stack, token)
		case "(":
			stack = append(stack, token)
		case ")":
			for len(stack) > 0 && stack[len(stack)-1] != "(" {
				rpn = append(rpn, stack[len(stack)-1])
				stack = stack[:len(stack)-1]
			}
			if len(stack) == 0 {
				return nil, errors.New("Несовпадение скобок")
			}
			stack = stack[:len(stack)-1]
		default:
			rpn = append(rpn, token)
		}
	}
	for len(stack) > 0 {
		top := stack[len(stack)-1]
		if top == "(" {
			return nil, errors.New("Несовпадение скобок")
		}
		rpn = append(rpn, top)
		stack = stack[:len(stack)-1]
	}
	return rpn, nil
}

func evaluateRPN(rpn []string) (float64, error) {
	var stack []float64
	for _, token := range rpn {
		switch token {
		case "+", "-", "*", "/":
			if len(stack) < 2 {
				return 0, errors.New("Ошибка вычисления: недостаточно операндов")
			}
			b, a := stack[len(stack)-1], stack[len(stack)-2]
			stack = stack[:len(stack)-2]
			switch token {
			case "+":
				stack = append(stack, a+b)
			case "-":
				stack = append(stack, a-b)
			case "*":
				stack = append(stack, a*b)
			case "/":
				if b == 0 {
					return 0, errors.New("Деление на ноль")
				}
				stack = append(stack, a/b)
			}
		default:
			num, err := strconv.ParseFloat(token, 64)
			if err != nil {
				return 0, errors.New("Ошибка преобразования числа")
			}
			stack = append(stack, num)
		}
	}
	if len(stack) != 1 {
		return 0, errors.New("Ошибка вычисления: неверное количество элементов на стеке")
	}
	return stack[0], nil
}

func main() {
	expression := " 3 + 5 * ( 2 - 4) /   2"
	result, err := Calc(expression)
	if err != nil {
		println("Ошибка:", err.Error())
	} else {
		println("Результат:", result)
	}
}
