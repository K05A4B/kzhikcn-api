package parser

import (
	"fmt"
	"kzhikcn/pkg/queryfilter/utils"

	"github.com/pkg/errors"
)

type SyntaxError struct {
	Err   string
	At    int
	Token *Token
}

func (s *SyntaxError) Error() string {
	if s.Token == nil {
		return fmt.Sprintf("syntax error: %s.", s.Err)
	}

	return fmt.Sprintf("syntax error: %s. (token \"%s\" at #%d)", s.Err, s.Token.Value, s.At)
}

type AstNode struct {
	Token *Token

	Left  *AstNode
	Right *AstNode
}

var operatorPrecedence = map[string]int{
	"!": 3,
	"=": 2, ">": 2, "<": 2, "~": 2, ">=": 2, "<=": 2, "!=": 2,
	"&": 1,
	"|": 0,
}

func isExpr(n *AstNode) bool {
	if n == nil {
		return false
	}

	if n.Token.Type != OPERATOR {
		return false
	}

	return n.Left != nil || n.Right != nil
}

// 判断左值优先级是否高于右值
func IsHigher(left, right string) bool {

	lp, ok := operatorPrecedence[left]
	if !ok {
		return false
	}

	rp, ok := operatorPrecedence[right]
	if !ok {
		return false
	}

	return lp > rp
}

// 隐式处理token优先级问题
// ConvertToRPN 采用 Shunting Yard 算法，将 Tokens 转换为 RPN（逆波兰表达式）
func TokensToRPN(tokens []Token) ([]Token, error) {
	output := utils.Stack[Token]{}
	operatorStack := utils.Stack[Token]{}
	parenCount := 0

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]

		if token.Type == PAREN && token.Value == "(" {
			parenCount++
			operatorStack.Push(token)
			continue
		}

		if token.Type == PAREN && token.Value == ")" {
			parenCount--
			if parenCount < 0 {
				return nil, &SyntaxError{"unmatched \")\"", i, &token}
			}

			for !operatorStack.IsEmpty() {
				top := operatorStack.Pop()

				if top.Type == PAREN && top.Value == "(" {
					continue
				}

				output.Push(top)
			}

			continue
		}

		if token.Type != OPERATOR {
			output.Push(token)
			continue
		}

		if IsHigher(operatorStack.Peek().Value, token.Value) {
			output.Push(operatorStack.Pop())
		}

		operatorStack.Push(token)
	}

	if parenCount != 0 {
		return nil, &SyntaxError{"unmatched parentheses", -1, nil}
	}

	for !operatorStack.IsEmpty() {
		top := operatorStack.Pop()

		if top.Type == PAREN {
			return nil, &SyntaxError{"mismatched parentheses", -1, &top}
		}

		output.Push(top)
	}

	return output.Copy(), nil
}

func RPNToAst(tokens []Token) (*AstNode, error) {
	stack := utils.Stack[*AstNode]{}

	for i := 0; i < len(tokens); i++ {
		token := tokens[i]

		if token.Type != OPERATOR {
			stack.Push(&AstNode{&token, nil, nil})
			continue
		}

		if stack.IsEmpty() {
			return nil, &SyntaxError{"the left side of the operator cannot be empty", i, &token}
		}

		if IsUnaryOperator(&token) {
			tree := &AstNode{&token, stack.Pop(), nil}

			stack.Push(tree)
			continue
		}

		right := stack.Pop()
		left := stack.Pop()

		tree := &AstNode{&token, left, right}

		err := checkBinaryExpr(tree)
		if err != nil {
			return nil, &SyntaxError{err.Error(), i, &token}
		}

		stack.Push(tree)
	}

	if stack.Len() != 1 {
		return nil, errors.Errorf("invalid tokens")
	}

	return stack.Pop(), nil
}

func checkBinaryExpr(ast *AstNode) error {
	token := ast.Token
	if token.Type != OPERATOR {
		return nil
	}

	if IsUnaryOperator(token) {
		return nil
	}

	left := ast.Left
	right := ast.Right

	if left == nil || right == nil {
		return errors.Errorf("both sides of the operator cannot be empty")
	}

	// 比较表达式的语法检查
	// 检查项
	// 	- 左侧必须是字段
	// 	- 右侧必须是值
	if IsComparisonOperator(token.Value) {
		if left.Token.Type != FIELD {
			return errors.New("the left side of the comparison operator must be a field")
		}

		if !IsTokenValue(right.Token.Type) {
			return errors.New("the right side of the comparison operator must be a value")
		}
	}

	// 逻辑表达式的语法检查（除去一元逻辑表达式）
	// 检查项:
	// 	- 检查左侧是不是表达式
	// 	- 检查右侧分支是不是表达式或者布尔值
	if IsLogicalOperator(token.Value) {
		if !isExpr(left) {
			return errors.New("the left side of the logical operator needs to be an expression")
		}

		if !isExpr(right) && right.Token.Type != BOOL {
			return errors.New("the right side of the logical operator needs to be an expression or a regular Boolean")
		}
	}

	return nil
}

func ParseToAST(str string) (*AstNode, error) {
	tokens, err := Tokenize(str)
	if err != nil {
		return nil, err
	}

	rpn, err := TokensToRPN(tokens)
	if err != nil {
		return nil, err
	}

	ast, err := RPNToAst(rpn)
	if err != nil {
		return nil, err
	}

	return ast, ValidateAST(ast)
}

// ValidateAST 检查AST的基本结构，确保逻辑运算符和NOT运算符正确
func ValidateAST(ast *AstNode) error {
	if ast == nil {
		return errors.New("AST is nil")
	}

	stack := utils.Stack[*AstNode]{}
	stack.Push(ast)

	for !stack.IsEmpty() {
		curr := stack.Pop()

		// NOT 运算符检查
		if IsUnaryOperator(curr.Token) {
			if curr.Left == nil {
				return errors.Errorf("NOT operator requires an operand")
			}
			stack.Push(curr.Left)
			continue
		}

		// 叶子节点检查
		if curr.Left == nil || curr.Right == nil {
			return errors.Errorf("invalid leaf node: '%s' missing field or value", curr.Token.Value)
		}
	}

	return nil
}
