package queryfilter

import (
	"fmt"
	"kzhikcn/pkg/queryfilter/parser"
	"kzhikcn/pkg/queryfilter/utils"

	"github.com/pkg/errors"
)

const (
	noTraversal uint8 = iota
	traversalLeft
	traversalFinish
)

var (
	ErrInvalidField = errors.New("invalid field")
)

func isNotOperator(node *parser.AstNode) bool {
	if node == nil {
		return false
	}

	return node.Token.Value == "!" && node.Token.Type == parser.OPERATOR
}

func toLogicKeyword(str string) string {
	switch str {
	case "|":
		return "OR"

	case "&":
		return "AND"

	default:
		return ""
	}
}

func ExprToSQL(expr string, wt WhiteList) (query string, vars []any, err error) {
	ast, err := parser.ParseToAST(expr)
	if err != nil {
		return "", nil, err
	}

	return AstToSQL(ast, wt)
}

func AstToSQL(ast *parser.AstNode, wt WhiteList) (query string, vars []any, err error) {
	vars = []any{}
	nodeStatus := map[*parser.AstNode]uint8{}
	stack := utils.Stack[*parser.AstNode]{}
	exprStack := utils.Stack[string]{}

	curr := ast

	for {
		if curr != nil {
			if parser.IsLogicalOperator(curr.Token.Value) && !isNotOperator(curr) {
				stack.Push(curr)
				nodeStatus[curr] = noTraversal
				curr = curr.Left
				continue
			}

			// 对于 ! 运算符是有bug的目前不知道如何解决
			template := "%s"
			if isNotOperator(curr) {
				template = "NOT %s"
				curr = curr.Left
			}

			operator := curr.Token.Value
			field := curr.Left
			value := curr.Right

			var valueParser ValueParser

			if wt != nil {
				var exist bool
				valueParser, exist = wt[field.Token.Value]
				if !exist {
					err = fmt.Errorf("%w: undefined field \"%s\"", ErrInvalidField, field.Token.Value)
					return
				}
			}

			if valueParser == nil {
				valueParser = ValueParserFunc(DefaultValueParser)
			}

			if operator == "~" {
				operator = "LIKE"
			}

			var goTypeValue any
			goTypeValue, err = valueParser.Parse(Token(*value.Token))
			if err != nil {
				err = errors.Wrap(err, "value parsing failed")
				return
			}

			vars = append(vars, goTypeValue)
			sqlExpr := fmt.Sprintf(template, fmt.Sprintf("%s %s ?", field.Token.Value, operator))
			exprStack.Push(sqlExpr)

			curr = nil
			continue
		}

		if stack.IsEmpty() {
			break
		}

		// 处理逻辑运算符节点
		parent := stack.Peek()
		parentStatus := nodeStatus[parent]

		if parentStatus == noTraversal {
			nodeStatus[parent] = traversalLeft
			curr = parent.Right
			continue
		}

		if parentStatus == traversalLeft {
			right := exprStack.Pop()
			left := exprStack.Pop()
			op := parent.Token.Value

			if parser.IsHigher(op, parent.Left.Token.Value) {
				left = fmt.Sprintf("(%s)", left)
			}

			if parser.IsHigher(op, parent.Right.Token.Value) {
				right = fmt.Sprintf("(%s)", right)
			}

			exprStack.Push(fmt.Sprintf("%s %s %s", left, toLogicKeyword(op), right))
			nodeStatus[parent] = traversalFinish
			stack.Pop()
			continue
		}
	}

	// 结果处理
	if !exprStack.IsEmpty() {
		query = exprStack.Pop()
	}

	return
}
