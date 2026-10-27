package middlewares

import (
	"fmt"
	"kzhikcn/pkg/config"
	"kzhikcn/pkg/hdl"
	"kzhikcn/pkg/log"
	"kzhikcn/server/common/authtoken"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"time"

	"github.com/sethvargo/go-limiter"
	limiterMemStore "github.com/sethvargo/go-limiter/memorystore"
)

const (
	// 配置缺失或非法时的回退配额，避免 limiter 初始化失败后为 nil。
	defaultRateTokens   = 100
	defaultRateInterval = time.Second
)

var ErrInvalidRemoteAddr = hdl.DefineError(500, "错误的远程地址", "system.httprate.parse_remote_addr_failed")

type httpRate struct {
	limiter   limiter.Store
	blackList []netip.Prefix
	apiKeys   map[string]struct{}
}

func HttpRate(conf *config.Config) func(http.Handler) http.Handler {
	r := &httpRate{
		apiKeys: make(map[string]struct{}),
	}

	r.preBuild(conf)

	return r.handler
}

func (hrate *httpRate) preBuild(c *config.Config) {
	for _, key := range c.HttpRate.HighQuotaKeys {
		// 空值会匹配所有未携带 X-High-Quota-Key 的请求，必须忽略。
		if key == "" {
			log.Warn("http_rate.high_quota_keys 包含空值，已忽略")
			continue
		}
		hrate.apiKeys[key] = struct{}{}
	}

	for _, addr := range c.HttpRate.BlackList {
		hrate.blackList = append(hrate.blackList, addr.ToPrefix())
	}

	tokens := c.HttpRate.LimitPerIP.Max
	interval := c.HttpRate.LimitPerIP.Window
	if tokens <= 0 || interval <= 0 {
		log.Warnf("tp_rate.limit_per_ip configuration is invalid (%d/%s), falling back to %d/%s",
			tokens, interval, defaultRateTokens, defaultRateInterval)
		tokens, interval = defaultRateTokens, defaultRateInterval
	}

	var err error
	hrate.limiter, err = limiterMemStore.New(&limiterMemStore.Config{
		Tokens:   uint64(tokens),
		Interval: interval,
	})
	if err != nil {
		// 参数已在上面兜底，正常不会失败；继续运行会因 nil limiter 在首个请求时 panic。
		panic(fmt.Errorf("Failed to initialize http_rate limiter: %w", err))
	}
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
			host = r.RemoteAddr
		}

		ip, err := netip.ParseAddr(host)
		if err != nil {
			host = "0.0.0.0"
			log.Error("parse remote addr failed: ", err)
			meta["failed_to_parse_remote_addr"] = r.RemoteAddr
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
