package queryfilter

import (
	"fmt"
	"kzhikcn/pkg/queryfilter/parser"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type TokenType parser.TokenType

const (
	UNKNOWN TokenType = 100 + iota
	STRING
	INT
	FLOAT
	BOOL
	PREDEFINED_VALUE
)

type Token parser.Token

func (t *Token) TypeOf() TokenType {
	switch parser.TokenType(t.Type) {
	case parser.STRING:
		return STRING

	case parser.INT:
		return INT

	case parser.FLOAT:
		return FLOAT

	case parser.BOOL:
		return BOOL

	case parser.PREDEFINED_VALUE:
		return PREDEFINED_VALUE

	default:
		return UNKNOWN
	}
}

func (t *Token) ConvertToGoType() (any, error) {
	token := parser.Token(*t)
	return parser.ConvertToGoType(&token)
}

type WhiteList map[string]ValueParser

func NewWhiteList() WhiteList {
	return WhiteList{}
}

func (w *WhiteList) ensureInitialized() {
	if *w == nil {
		*w = make(WhiteList)
	}
}

func (w WhiteList) Add(fields ...string) WhiteList {
	for _, field := range fields {
		w.AddWithParser(field, nil)
	}
	return w
}

func (w WhiteList) AddWithParser(field string, valParser ValueParser) WhiteList {
	w.ensureInitialized()
	w[field] = valParser

	return w
}

type ValueParser interface {
	Parse(val Token) (any, error)
}

type ValueParserFunc func(Token) (any, error)

func (fn ValueParserFunc) Parse(token Token) (any, error) {
	return fn(token)
}

func DefaultValueParser(token Token) (any, error) {
	newToken := parser.Token(token)
	return parser.ConvertToGoType(&newToken)
}

// 将值解析成时间格式
func TimeValueParser() ValueParser {

	paddingZero := func(t int) string {
		if t < 10 {
			return fmt.Sprintf("0%d", t)
		}

		return strconv.Itoa(t)
	}

	paddingZeroArr := func(arr []string) []string {
		res := []string{}

		for i := 0; i < len(arr); i++ {
			v, _ := strconv.Atoi(arr[i])
			res = append(res, paddingZero(v))
		}

		return res
	}

	return ValueParserFunc(func(t Token) (any, error) {
		unsupportedErr := errors.Errorf("unsupported value type: expected 'yyyy-mm-dd hh:mm:ss' or Unix timestamp, got %s", t.Value)
		if t.Type == parser.INT {
			val, err := t.ConvertToGoType()
			if err != nil {
				return nil, err
			}

			return time.Unix(val.(int64), 0), nil
		}

		if t.Type != parser.STRING {
			return nil, unsupportedErr
		}

		val := t.Value
		parts := strings.Split(val, " ")
		if len(parts) == 2 {
			return time.Parse(time.DateTime, t.Value)
		}

		if len(parts) == 1 {
			datePart := strings.Split(parts[0], "-")
			timePart := strings.Split(parts[0], ":")

			if len(datePart) > 1 {
				partsLen := len(datePart)
				format := "2006-01-02"
				if partsLen == 2 {
					format = "2006-01"
				}

				datePart = paddingZeroArr(datePart)

				return time.Parse(format, strings.Join(datePart, "-"))
			}

			if len(timePart) > 1 {
				partsLen := len(timePart)

				format := "15:04:05"

				if partsLen == 2 {
					format = "15:04"
				}

				if format == "" {
					return nil, unsupportedErr
				}

				timePart = paddingZeroArr(timePart)

				now := time.Now()
				value := fmt.Sprintf("%d-%s-%s %s", now.Year(), paddingZero(int(now.Month())), paddingZero(now.Day()), strings.Join(timePart, ":"))

				t, err := time.Parse(fmt.Sprintf("2006-01-02 %s", format), value)

				return t, err
			}
		}

		return nil, unsupportedErr
	})
}
