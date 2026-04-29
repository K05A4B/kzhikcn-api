package parser

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestTokensToRPN(t *testing.T) {
	tokens, _ := Tokenize("(age > 18 & city = \"New York\") | (score = 0 & height < 180) | !(country = \"US\" & score = 0)")
	rpn, err := TokensToRPN(tokens)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println(rpn)

	for _, token := range rpn {
		fmt.Print(token.Value + " ")
	}

	fmt.Println()
}

func TestRpnToAst(t *testing.T) {
	tokens, _ := Tokenize(`age>18 & city="New York" &  (height<180) & !(ip_location ~ "US%" | score=null)`)
	rpn, _ := TokensToRPN(tokens)
	ast, err := RPNToAst(rpn)
	if err != nil {
		t.Error(err)
		return
	}

	data, _ := json.Marshal(ast)
	t.Log(string(data))
}

func TestRpnToAstCheckSyntax(t *testing.T) {
	invalidExpressions := []string{
		"!18",
		"!\"text\"",
		"!(age > 18 & \"NY\")",
		"age > 18 &",
		"& age < 30",
		"age > 18 |",
		"|",
		"&",
		"!",
		"= 18",
		"age >",
		"> 30",
		"\"name\" > 18",
		"city > true",
		"(age > 18",
		"age > 18)",
		"!(city = \"NY\"",
		"()",
		"!( )",
		"(age > 18 & )",
		"age >>> 18",
		"!>> age",
		"age > > 18",
		"age =",
		"height <",
		"!(age > 18 & (city = \"NY\" & ))",
		"!((age > 18) & !())",
		"!(city = \"NY\" & height >)",
	}

	for _, expr := range invalidExpressions {
		t.Run(expr, func(t *testing.T) {
			_, err := ParseToAST(expr)
			if err == nil {
				t.Error("错误未检出")
				return
			}

			t.Log("错误已检出:", err)
		})
	}
}

// func TestParseToTree(t *testing.T) {
// 	res, errs := ParseToTree("(age > 18 & city = \"New York\") | (score = 0 & height < 180) | !(country = \"US\" & score = 0)")
// 	if errs != nil {
// 		t.Error(errs)
// 		return
// 	}

// 	data, _ := json.Marshal(res)
// 	fmt.Println(string(data))
// }
