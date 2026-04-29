package httputil

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

type CacheDirective interface {
	Directive() string
}

type CacheDirectiveFunc func() string

func (c CacheDirectiveFunc) Directive() string {
	return c()
}

func CachePublic() CacheDirective {
	return CacheDirectiveFunc(func() string {
		return "public"
	})
}

func NoCache() CacheDirective {
	return CacheDirectiveFunc(func() string {
		return "no-cache"
	})
}

func CacheImmutable() CacheDirective {
	return CacheDirectiveFunc(func() string {
		return "immutable"
	})
}

func CacheMaxAge(d time.Duration) CacheDirective {
	return CacheDirectiveFunc(func() string {
		return fmt.Sprintf("max-age=%d", int64(math.Ceil(d.Seconds())))
	})
}

func ApplyCacheControl(w http.ResponseWriter, cds ...CacheDirective) {
	directives := []string{}

	for _, cd := range cds {

		directives = append(directives, cd.Directive())
	}

	w.Header().Set("Cache-Control", strings.Join(directives, ", "))
}
