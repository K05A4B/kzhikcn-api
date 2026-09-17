package cli

import (
	"testing"

	pkgconfig "kzhikcn/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validConfig() pkgconfig.Config {
	var conf pkgconfig.Config
	conf.Database.Driver = "sqlite3"
	conf.Database.Dsn = ":memory:"
	conf.Storage.Provider = "local"
	conf.Storage.Articles.BasePath = "./data/articles"
	conf.Cache.Provider = "local"
	conf.Cache.Local.Dir = "./sys/cache"
	conf.Auth.JWT.Secret = "test-secret"
	conf.Log.LogLevel = "info"
	return conf
}

func TestValidateConfig_Valid(t *testing.T) {
	conf := validConfig()
	assert.Empty(t, validateConfig(&conf))
}

func TestValidateConfig_EmptyConfig(t *testing.T) {
	var conf pkgconfig.Config
	problems := validateConfig(&conf)

	assert.Contains(t, problems, "db.driver 不能为空")
	assert.Contains(t, problems, "db.dsn 不能为空")
	assert.Contains(t, problems, "storage.provider 不能为空")
	assert.Contains(t, problems, "cache.provider 不能为空")
	assert.Contains(t, problems, "auth.jwt.secret 不能为空")
}

func TestValidateConfig_UnresolvedEnv(t *testing.T) {
	conf := validConfig()
	conf.Auth.JWT.Secret = "${JWT_SECRET}"
	conf.MRR.BaseUrl = "${WEBSITE_URL}"

	problems := validateConfig(&conf)
	assert.Contains(t, problems, "auth.jwt.secret 存在未解析的环境变量占位符")
	assert.Contains(t, problems, "machine_readable_resources.base_url 存在未解析的环境变量占位符")
}

func TestValidateConfig_InvalidValues(t *testing.T) {
	conf := validConfig()
	conf.Database.Driver = "postgres"
	conf.Cache.Provider = "memcached"
	conf.Log.LogLevel = "verbose"

	problems := validateConfig(&conf)
	assert.Contains(t, problems, "db.driver 不受支持: postgres")
	assert.Contains(t, problems, "cache.provider 不受支持: memcached")
	assert.Contains(t, problems, "log.log_level 非法: verbose")
}

func TestMaskConfig(t *testing.T) {
	conf := validConfig()
	conf.Cache.Redis.Password = "redis-password"

	masked := maskConfig(&conf)

	assert.Equal(t, maskedValue, masked.Auth.JWT.Secret)
	assert.Equal(t, maskedValue, masked.Cache.Redis.Password)

	// 原配置不应被修改
	require.Equal(t, "test-secret", conf.Auth.JWT.Secret)
	require.Equal(t, "redis-password", conf.Cache.Redis.Password)
}
