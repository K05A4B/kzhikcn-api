package httputil

import (
	"context"
	"fmt"
	"kzhikcn/pkg/data"
	"kzhikcn/pkg/log"
	"kzhikcn/pkg/queryfilter"
	"kzhikcn/pkg/utils"
	"kzhikcn/server/common/hdl"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type contextMark string

func InvalidExpression(err error) *hdl.HandlerError {
	return hdl.Error(400, err.Error(), err, "system.expr.invalid")
}

func RedirectTo(url string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusPermanentRedirect)
		w.Header().Set("Location", url)
	}
}

func QueryString(r *http.Request, key, defaultValue string) string {
	query := r.URL.Query()

	value := query.Get(key)
	if value == "" {
		value = defaultValue
	}

	return value
}

func QueryBool(r *http.Request, key string, defaultValue bool) bool {
	def := "false"
	if defaultValue {
		def = "true"
	}

	val, _ := strconv.ParseBool(QueryString(r, key, def))
	return val
}

func QueryInt(r *http.Request, key string, defaultValue int) int {
	res, _ := strconv.Atoi(QueryString(r, key, strconv.Itoa(defaultValue)))
	return res
}

func WithMarks(r *http.Request, marks ...string) *http.Request {
	ctx := r.Context()
	for _, mark := range marks {
		ctx = context.WithValue(ctx, contextMark(mark), true)
	}

	return r.WithContext(ctx)
}

func HasMark(r *http.Request, name string) bool {
	val := r.Context().Value(contextMark(name))

	if val == nil {
		return false
	}

	res, ok := val.(bool)

	if !ok {
		return false
	}

	return res
}

func UseExpression(r *http.Request, wt queryfilter.WhiteList) (modifier data.QueryModifier, err error) {
	query := ""
	vars := []any{}
	hasExpr := false

	expr := QueryString(r, "expr", "")
	if strings.TrimSpace(expr) != "" {
		hasExpr = true

		query, vars, err = queryfilter.ExprToSQL(expr, wt)
		if err != nil {
			return
		}
	}

	modifier = func(tx *gorm.DB) *gorm.DB {
		if !hasExpr {
			return tx
		}

		return tx.Where(query, vars...)
	}

	return
}

func ApplyOrderBy(r *http.Request, defaultPolicy string, tx *gorm.DB, policies map[string]string) *gorm.DB {
	orderBy := QueryString(r, "orderBy", defaultPolicy)

	policy, ok := policies[orderBy]

	if !ok {
		policy = policies[defaultPolicy]
	}

	return tx.Order(policy)
}

// ExtendOrderByWith 添加 ":desc" 版本的排序规则，生成包含降序选项的新 map。
// 例如：{"createdAt": "created_at"} -> {"createdAt": "created_at", "createdAt:desc": "created_at DESC"}
func ExtendOrderByWithDesc(policies map[string]string) map[string]string {
	copied := utils.CopyMap(policies)

	for key, val := range policies {
		copied[(key + ":desc")] = val + " DESC"
	}

	return copied
}

func Pagination(r *http.Request, defaultLimit, maxLimit int) (page, limit int) {
	page = QueryInt(r, "page", 1)
	limit = QueryInt(r, "limit", defaultLimit)

	if page <= 0 {
		page = 1
	}

	if limit < 0 {
		limit = defaultLimit
	}

	if limit > maxLimit {
		limit = maxLimit
	}

	return
}

func ApplyPagination(r *http.Request, defaultLimit, maxLimit int, tx *gorm.DB) *gorm.DB {
	page, limit := Pagination(r, defaultLimit, maxLimit)
	tx = tx.Limit(limit)
	tx = tx.Offset((page - 1) * limit)
	return tx
}

func SetTotal(resp *hdl.Response, model any, modifiers ...data.QueryModifier) {
	total, err := data.Total(model, modifiers...)
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.With("caller", fmt.Sprintf("%s:%d", file, line)).Error(err)

		resp.Message = "Warning: Failed to count the total number"
		return
	}

	if resp.Meta == nil {
		resp.Meta = make(map[string]any)
	}

	resp.Meta["total"] = total
}

func GetPassID(r *http.Request) string {
	passId := ""

	passCookie, err := r.Cookie("pass_id")
	if passCookie != nil && err == nil {
		passId = passCookie.Value
	} else {
		passId = r.Header.Get("x-pass-id")
	}

	return passId
}

func Accepts(r *http.Request, mime ...string) bool {
	accept := r.Header.Get("Accept")

	for _, m := range mime {
		if strings.Contains(accept, m) {
			return true
		}
	}

	return false
}
