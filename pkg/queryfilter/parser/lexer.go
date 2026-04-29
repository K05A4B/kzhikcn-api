package parser

import (
	"slices"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type TokenType int

const (
	UNKNOWN_TYPE TokenType = -1

	FIELD TokenType = iota
	OPERATOR
	PAREN
	STRING
	INT
	FLOAT
	BOOL
	PREDEFINED_VALUE
)

var predefinedValues = []string{"null"}

type Token struct {
	Type  TokenType
	Value string
}

func ConvertToGoType(token *Token) (any, error) {
	if token == nil || !IsTokenValue(token.Type) {
		return nil, errors.Errorf("token value must match its expected type (%s)", token.Value)
	}

	val := token.Value

	switch token.Type {
	case INT:
		return strconv.ParseInt(val, 10, 64)

	case FLOAT:
		return strconv.ParseFloat(val, 64)

	case STRING:
		return val, nil

	case BOOL:
		return strconv.ParseBool(val)

	case PREDEFINED_VALUE:
		return convertPredefinedValue(token)

	default:
		return nil, errors.Errorf("unsupported token type: %v, cannot convert to Go type", token.Type)
	}
}

func IsComparisonOperator(s string) bool {
	return s == ">" || s == "<" || s == "=" || s == ">=" || s == "<=" || s == "!=" || s == "~"
}

func IsLogicalOperator(s string) bool {
	return s == "&" || s == "|" || s == "!"
}

func IsTokenValue(t TokenType) bool {
	return t == BOOL || t == INT || t == STRING || t == PREDEFINED_VALUE || t == FLOAT
}

func IsUnaryOperator(t *Token) bool {
	if t == nil {
		return false
	}

	if t.Type != OPERATOR {
		return false
	}

	return t.Value == "!"
}

func convertPredefinedValue(token *Token) (any, error) {
	if token == nil || token.Type != PREDEFINED_VALUE {
		return nil, errors.Errorf("token type must be PREDEFINED_VALUE")
	}

	switch token.Value {
	case "null":
		return nil, nil
	}

	return nil, errors.Errorf("undefined predefine value: %s", token.Value)
}

// 判断某个token的原始值是不是操作符
func isOperator(str string) bool {
	_, ok := operatorPrecedence[str]
	return ok
}

func isEmptyToken(t rune) bool {
	return strings.TrimSpace(string(t)) == ""
}

func isPredefinedValue(str string) bool {
	return slices.Contains(predefinedValues, str)
}

func invalidToken(s []rune, i int) error {
	if i < 0 || i >= len(s) {
		return errors.Errorf("(invalid token) unexpected end of input at column %d (%s)", i, s)
	}

	start := max(i-5, 0)
	end := min(i+5, len(s))

	context := s[start:end]
	return errors.Errorf("(invalid token) '%s' at column %d, context: ...%s...", string(s[i]), i, context)
}

func tokenizeNumber(s []rune, begin int) (n string, nextBegin int, isFloat bool, err error) {
	hasDot := false

	invalidNumberError := func(i int) error {
		return invalidToken(s, i)
	}

	for i := begin; i < len(s); i++ {
		c := s[i]
		cs := string(c)

		if isEmptyToken(c) {
			continue
		}

		if (c == '.' && hasDot) || (c == '+' || c == '-' && i != begin) {
			return n, i, false, invalidNumberError(i)
		}

		if c == '.' {
			hasDot = true
		}

		if c >= '0' && c <= '9' {
			n += cs
			continue
		}

		if isOperator(cs) || c == '(' || c == ')' {
			return n, i - 1, hasDot, nil
		}

		return n, i, false, invalidNumberError(i)
	}

	return n, len(s) - 1, hasDot, nil
}

func tokenizeString(s []rune, begin int) (value string, nextBegin int, err error) {
	literal := rune(s[begin])

	errNotTerminated := errors.New("string literal not terminated")

	if len(s)-1 == begin {
		return "", len(s) - 1, errNotTerminated
	}

	for i := begin + 1; i < len(s); i++ {
		c := s[i]

		if c == literal {
			return value, i, nil
		}

		value += string(c)
	}

	return "", len(s) - 1, errNotTerminated
}

func tokenizeUncertain(s []rune) (*Token, error) {
	s = []rune(strings.TrimSpace(string(s)))
	str := string(s)

	if len(s) == 0 {
		return nil, nil
	}

	_, err := strconv.ParseBool(str)
	if err == nil {
		return &Token{BOOL, str}, nil
	}

	if isPredefinedValue(string(s)) {
		return &Token{PREDEFINED_VALUE, str}, nil
	}

	for _, c := range s {
		if isEmptyToken(c) {
			return nil, invalidToken(s, len(s))
		}
	}

	return &Token{FIELD, str}, nil
}

// 分词器
// 用来分词字符串
func Tokenize(str string) ([]Token, error) {
	tokens := []Token{}
	s := []rune(str)

	uncertain := []rune{}

	appendToken := func(t Token) error {
		// 在遇到已知的token类型时尝试解析不确定的token
		if len(uncertain) != 0 {
			parsedToken, err := tokenizeUncertain(uncertain)
			if err != nil {
				return err
			}

			if parsedToken != nil {
				tokens = append(tokens, *parsedToken)
				uncertain = []rune{}
			}
		}

		tokens = append(tokens, t)
		return nil
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		cs := string(c)

		if c == '(' || c == ')' {
			if err := appendToken(Token{PAREN, cs}); err != nil {
				return nil, err
			}

			continue
		}

		if c == '"' || c == '\'' || c == '`' {
			value, next, err := tokenizeString(s, i)
			if err != nil {
				return nil, err
			}

			i = next

			if err := appendToken(Token{STRING, value}); err != nil {
				return nil, err
			}

			continue
		}

		if c >= '0' && c <= '9' || c == '+' || c == '-' {
			number, next, isFloat, err := tokenizeNumber(s, i)
			if err != nil {
				return nil, err
			}

			i = next

			tokenType := INT
			if isFloat {
				tokenType = FLOAT
			}

			if err := appendToken(Token{tokenType, number}); err != nil {
				return nil, err
			}

			continue
		}

		if isOperator(cs) {
			t := Token{OPERATOR, cs}

			if (c == '>' || c == '<' || c == '!' || c == '=') && i+1 < len(s) && s[i+1] == '=' {
				// == 解析为 =
				if c != '=' {
					t.Value += "="
				}

				i++
			}

			if err := appendToken(t); err != nil {
				return nil, err
			}

			continue
		}

		uncertain = append(uncertain, c)
	}

	if len(uncertain) != 0 {
		parsedToken, err := tokenizeUncertain(uncertain)
		if err != nil {
			return nil, err
		} else if parsedToken != nil {
			tokens = append(tokens, *parsedToken)
		}
	}

	return tokens, nil
}
