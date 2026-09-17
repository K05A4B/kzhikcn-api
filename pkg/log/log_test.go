package log

import (
	"bytes"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetLogger(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		_ = Configure(Options{Level: "info", Console: true})
	})
}

func TestDefaultLoggerAvailable(t *testing.T) {
	assert.NotNil(t, GetLogger())
	assert.NotNil(t, GetZeroLogger())
}

func TestConfigure_InvalidLevel(t *testing.T) {
	resetLogger(t)

	require.Error(t, Configure(Options{Level: "verbose"}))
}

func TestConfigure_LevelFiltering(t *testing.T) {
	resetLogger(t)

	var buf bytes.Buffer
	require.NoError(t, Configure(Options{Level: "warn", Console: false, Writers: []io.Writer{&buf}}))

	Info("filtered-message")
	Warn("kept-message")

	out := buf.String()
	assert.NotContains(t, out, "filtered-message")
	assert.Contains(t, out, "kept-message")
}

func TestConfigure_NilWriterIgnored(t *testing.T) {
	resetLogger(t)

	var buf bytes.Buffer
	require.NoError(t, Configure(Options{Level: "info", Console: false, Writers: []io.Writer{nil, &buf}}))

	Info("hello")

	assert.Contains(t, buf.String(), "hello")
}

func TestSetLevel(t *testing.T) {
	resetLogger(t)

	var buf bytes.Buffer
	require.NoError(t, Configure(Options{Level: "error", Console: false, Writers: []io.Writer{&buf}}))

	Info("hidden")

	require.NoError(t, SetLevel("debug"))
	Debug("shown")

	out := buf.String()
	assert.NotContains(t, out, "hidden")
	assert.Contains(t, out, "shown")
}

func TestSetLevel_Invalid(t *testing.T) {
	resetLogger(t)

	require.Error(t, SetLevel("verbose"))
}

func TestError_AttachesCaller(t *testing.T) {
	resetLogger(t)

	var buf bytes.Buffer
	require.NoError(t, Configure(Options{Level: "error", Console: false, Writers: []io.Writer{&buf}}))

	Error("package-level")
	GetLogger().Error("method-level")

	out := buf.String()
	assert.Contains(t, out, `"caller"`)
	assert.Contains(t, out, "log_test.go")
	assert.True(t, strings.Count(out, "log_test.go") >= 2, "两级入口都应带 caller: %s", out)
}

func TestConfigure_Concurrent(t *testing.T) {
	resetLogger(t)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)

		go func() {
			defer wg.Done()
			_ = Configure(Options{Level: "info", Console: false, Writers: []io.Writer{io.Discard}})
		}()

		go func() {
			defer wg.Done()
			_ = SetLevel("error")
		}()

		go func() {
			defer wg.Done()
			GetLogger().Info("concurrent")
		}()
	}

	wg.Wait()
}
