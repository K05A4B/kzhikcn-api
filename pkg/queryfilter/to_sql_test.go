package queryfilter

import (
	"testing"
)

func TestExprToSQL(t *testing.T) {
	query, vars, err := ExprToSQL(`title = "测试页面" | description = '关于页面'`, nil)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log("Sql: ", query)
	t.Log("Vars: ", vars)
}

func TestExprToSQLWithWt(t *testing.T) {
	wt := WhiteList{}

	// 添加白名单和类型解析器
	wt.AddWithParser("age", ValueParserFunc(func(t Token) (any, error) {
		if t.Value == "18" {
			return "Test111", nil
		}

		return t.ConvertToGoType()
	})).
		Add("city", "ip_location", "score", "height")

	query, vars, err := ExprToSQL(`age>18 & city="New York" & height<180 & (ip_location ~ "US%" | score=null)`, wt)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log("Sql: ", query)
	t.Log("Vars: ", vars)
}
