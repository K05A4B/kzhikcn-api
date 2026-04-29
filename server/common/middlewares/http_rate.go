package middlewares

import (
	"kzhikcn/pkg/config"
	"kzhikcn/server/common/authtoken"
	"kzhikcn/server/common/hdl"
	"net"
	"net/http"
	"net/netip"
	"strconv"

	"github.com/sethvargo/go-limiter"
	limiterMemStore "github.com/sethvargo/go-limiter/memorystore"
)

var ErrInvalidRemoteAddr = hdl.DefineError(500, "错误的远程地址", "system.httprate.parse_remote_addr_failed")

type httpRate struct {
	limiter   limiter.Store
	blackList []netip.Prefix
	apiKeys   map[string]struct{}
}

func HttpRate() func(http.Handler) http.Handler {
	r := &httpRate{
		apiKeys: make(map[string]struct{}),
	}

	if config.Conf() == nil {
		config.HookLoaded(r.preBuild)
		return r.handler
	}

	r.preBuild(config.Conf())

	return r.handler
}

func (hrate *httpRate) preBuild(c *config.Config) {
	for _, key := range c.HttpRate.HighQuotaKeys {
		hrate.apiKeys[key.String()] = struct{}{}
	}

	for _, addr := range c.HttpRate.BlackList {
		hrate.blackList = append(hrate.blackList, addr.ToPrefix())
	}

	hrate.limiter, _ = limiterMemStore.New(&limiterMemStore.Config{
		Tokens:   uint64(c.HttpRate.LimitPerIP.Max),
		Interval: c.HttpRate.LimitPerIP.Window,
	})

}

func (hrate *httpRate) handler(h http.Handler) http.Handler {
	return hdl.Middleware(func(w http.ResponseWriter, r *http.Request, meta map[string]any) error {

		// 对于持有Token并校验通过的请求，直接放行
		claims := authtoken.GetClaims(r.Context())
		if claims != nil {
			h.ServeHTTP(w, r)
			return nil
		}

		// 提供了高配额密钥并校验通过的请求，直接放行
		key := r.Header.Get("X-High-Quota-Key")
		_, ok := hrate.apiKeys[key]
		if ok {
			h.ServeHTTP(w, r)
			return nil
		}

		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return ErrInvalidRemoteAddr
		}

		ip, err := netip.ParseAddr(host)
		if err != nil {
			return ErrInvalidRemoteAddr
		}

		for _, addr := range hrate.blackList {
			if addr.Contains(ip) {
				return hdl.Error(403, "Forbidden", nil, "system.httprate.forbidden")
			}
		}

		tokens, remaining, reset, ok, err := hrate.limiter.Take(r.Context(), host)
		if err != nil {
			return hdl.Error(500, "拿取令牌失败", err, "system.httprate.take_failed")
		}

		w.Header().Set("X-RateLimit-Limit", strconv.FormatUint(uint64(tokens), 10))
		w.Header().Set("X-RateLimit-Remaining", strconv.FormatUint(uint64(remaining), 10))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatUint(reset, 10))

		if !ok {
			meta["remaining"] = remaining
			meta["reset"] = reset
			meta["limit"] = tokens
			return hdl.Error(429, "Too many requests", nil, "system.httprate.too_many_requests")
		}

		h.ServeHTTP(w, r)
		return nil
	})
}
