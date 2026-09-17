package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"kzhikcn/internal/cli"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runCLI 以给定的参数运行一份 AppCli 副本，返回标准输出与错误。
func runCLI(t *testing.T, configFile string, args ...string) (string, error) {
	t.Helper()

	var buf bytes.Buffer

	app := cli.AppCli
	app.Writer = &buf
	app.ErrWriter = &buf

	full := append([]string{"kzhikcn", "-c", configFile}, args...)
	err := app.Run(full)

	return buf.String(), err
}

func writeTestConfig(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yml")

	content := fmt.Sprintf(`storage:
  provider: local
  articles:
    base_path: %s
cache:
  provider: local
  local:
    dir: %s
    value_log_file_size: 64mb
    mem_table_size: 8mb
db:
  driver: sqlite3
  dsn: "file:%s?_foreign_keys=on"
auth:
  jwt:
    secret: test-secret
    expiry: 72h
machine_readable_resources:
  base_url: https://example.com
log:
  enable: false
  log_level: error
`,
		filepath.ToSlash(filepath.Join(dir, "articles")),
		filepath.ToSlash(filepath.Join(dir, "cache")),
		filepath.ToSlash(filepath.Join(dir, "test.db")),
	)

	require.NoError(t, os.WriteFile(configFile, []byte(content), 0o600))
	return configFile
}

func TestAdminLifecycle(t *testing.T) {
	configFile := writeTestConfig(t)

	out, err := runCLI(t, configFile, "admin", "add", "-n", "tester", "-p", "secret123")
	require.NoError(t, err)
	assert.Contains(t, out, "添加成功")

	// 重复添加应失败
	_, err = runCLI(t, configFile, "admin", "add", "-n", "tester", "-p", "secret123")
	require.Error(t, err)

	// find 输出 JSON
	out, err = runCLI(t, configFile, "admin", "find", "-n", "tester", "--format", "json")
	require.NoError(t, err)

	var view struct {
		Username  string `json:"username"`
		EnableMFA bool   `json:"enableMFA"`
	}
	require.NoError(t, json.Unmarshal([]byte(out), &view))
	assert.Equal(t, "tester", view.Username)
	assert.False(t, view.EnableMFA)

	// list 至少包含默认管理员与 tester
	out, err = runCLI(t, configFile, "admin", "list", "--format", "json")
	require.NoError(t, err)

	var list []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &list))
	assert.GreaterOrEqual(t, len(list), 2)

	// 生成 MFA 并启用
	out, err = runCLI(t, configFile, "admin", "mfa", "-n", "tester", "--enable")
	require.NoError(t, err)
	assert.Contains(t, out, "otpauth URL")

	out, err = runCLI(t, configFile, "admin", "find", "-n", "tester", "--format", "json")
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal([]byte(out), &view))
	assert.True(t, view.EnableMFA)

	// 修改密码
	_, err = runCLI(t, configFile, "admin", "passwd", "-n", "tester", "-p", "newsecret")
	require.NoError(t, err)

	// 删除
	out, err = runCLI(t, configFile, "admin", "delete", "-n", "tester", "--yes")
	require.NoError(t, err)
	assert.Contains(t, out, "已删除")

	// 删除后再查询应失败
	_, err = runCLI(t, configFile, "admin", "find", "-n", "tester", "--format", "json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "没有找到管理员")
}

func TestAdminAdd_ShortPasswordRejected(t *testing.T) {
	configFile := writeTestConfig(t)

	_, err := runCLI(t, configFile, "admin", "add", "-n", "shorty", "-p", "123")
	require.Error(t, err)
}

func TestConfigCheck(t *testing.T) {
	configFile := writeTestConfig(t)

	out, err := runCLI(t, configFile, "config", "check")
	require.NoError(t, err)
	assert.Contains(t, out, "校验通过")
}

func TestConfigShow_MasksSecrets(t *testing.T) {
	configFile := writeTestConfig(t)

	out, err := runCLI(t, configFile, "config", "show", "--format", "json")
	require.NoError(t, err)

	var generic map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &generic))

	auth, ok := generic["auth"].(map[string]any)
	require.True(t, ok)
	jwt, ok := auth["jwt"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "******", jwt["secret"])
}

func TestMigrateCommand(t *testing.T) {
	configFile := writeTestConfig(t)

	out, err := runCLI(t, configFile, "migrate")
	require.NoError(t, err)
	assert.Contains(t, out, "数据库迁移完成")
}

func TestVersionFlag(t *testing.T) {
	var buf bytes.Buffer

	app := cli.AppCli
	app.Writer = &buf

	require.NoError(t, app.Run([]string{"kzhikcn", "--version"}))
	assert.NotEmpty(t, buf.String())
}
