package config_test

import (
	"kzhikcn/pkg/config"
	"net/netip"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestConfigLoad(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Join(filepath.Dir(file), "testdata")

	t.Setenv("WEBSITE_URL", "https://www.kzhik.cn")
	t.Setenv("HTTP_RATE_BLACK_LIST1", "123.231.128.0/23")
	t.Setenv("WEBHOOK_URL", "http://localhost:9000/hook")
	t.Setenv("WEBHOOK_TOKEN", "secret-token")

	c, err := config.LoadConfigFromFile(filepath.Join(dir, "test_config.yml"))
	if err != nil {
		t.Fatal(err)
	}

	// 测试普通字段环境变量读取情况
	// 预期 ${WEBSITE_URL} 被替换为 https://www.kzhik.cn
	if c.MRR.BaseUrl != "https://www.kzhik.cn" {
		t.Fatalf("test1 failed, env value: %s", os.Getenv("WEBSITE_URL"))
	}

	// 测试没有某个环境变量时的情况
	// 预期 ${JWT_SECRET} 未被替换
	if c.Auth.JWT.Secret != "${JWT_SECRET}" {
		t.Fatalf("test2 failed, env value: %s", os.Getenv("JWT_SECRET"))
	}

	// 测试列表配置项环境变量读取情况
	// 预期 ${HTTP_RATE_BLACK_LIST1} 被替换为 123.231.128.0/23
	ip123, _ := netip.ParseAddr("123.231.129.1")
	if !c.HttpRate.BlackList[0].ToPrefix().Contains(ip123) {
		t.Fatalf("test3 failed, env value: %v", os.Getenv("HTTP_RATE_BLACK_LIST1"))
	}

	// 测试其他类型的配置是否会被影响
	// 预期 http_rate.limit_per_ip 被解析为 RateLimit 且值正确
	if c.HttpRate.LimitPerIP.Max != 100 || c.HttpRate.LimitPerIP.Window != time.Second {
		t.Fatalf("test4 failed, limit_per_ip: %v/%v", c.HttpRate.LimitPerIP.Max, c.HttpRate.LimitPerIP.Window)
	}

	// 测试事件配置解析
	// 预期 event_dispatcher 被解析为对应字段
	if c.EventDispatcher.Timeout.Duration() != 3*time.Second {
		t.Fatalf("test5 failed, event_dispatcher.timeout: %v", c.EventDispatcher.Timeout.Duration())
	}
	if c.EventDispatcher.Workers != 2 || c.EventDispatcher.QueueSize != 8 {
		t.Fatalf("test5b failed, event_dispatcher: %+v", c.EventDispatcher)
	}

	// 预期 events 列表被完整解析，且 entry 中的环境变量被替换
	if len(c.Events) != 2 {
		t.Fatalf("test6 failed, events length: %d", len(c.Events))
	}
	if c.Events[0].Name != "T1" || c.Events[0].On != "article.created" || c.Events[0].Type != "webhook" {
		t.Fatalf("test7 failed, events[0]: %+v", c.Events[0])
	}
	if c.Events[0].Entry != "http://localhost:9000/hook" {
		t.Fatalf("test8 failed, events[0].entry: %s", c.Events[0].Entry)
	}
	if c.Events[1].Name != "T2" || c.Events[1].On != "auth.login_success" || c.Events[1].Type != "command" {
		t.Fatalf("test9 failed, events[1]: %+v", c.Events[1])
	}

	// 预期 events[].async 被解析：显式 true，缺省为 nil（默认异步）
	if c.Events[0].Async == nil || !*c.Events[0].Async {
		t.Fatalf("test10 failed, events[0].async: %v", c.Events[0].Async)
	}
	if c.Events[1].Async != nil {
		t.Fatalf("test11 failed, events[1].async: %v", c.Events[1].Async)
	}

	// 预期事件级 timeout 与 headers 被解析，且 header 值中的环境变量被替换
	if c.Events[0].Timeout.Duration() != 7*time.Second {
		t.Fatalf("test12 failed, events[0].timeout: %v", c.Events[0].Timeout.Duration())
	}
	if c.Events[0].Headers["Authorization"] != "Bearer secret-token" {
		t.Fatalf("test13 failed, events[0].headers.Authorization: %q", c.Events[0].Headers["Authorization"])
	}
	if c.Events[0].Headers["X-Source"] != "kzhikcn" {
		t.Fatalf("test14 failed, events[0].headers.X-Source: %q", c.Events[0].Headers["X-Source"])
	}
	if len(c.Events[1].Headers) != 0 {
		t.Fatalf("test15 failed, events[1].headers: %+v", c.Events[1].Headers)
	}
}
