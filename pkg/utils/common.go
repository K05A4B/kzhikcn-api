package utils

import (
	"crypto/rand"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func SafePath(base, unsafePath string) (target string, isSafe bool) {
	base, err := filepath.Abs(base)
	if err != nil {
		return "", false
	}
	base = filepath.Clean(base)

	target = filepath.Join(base, filepath.Clean(unsafePath))
	target, err = filepath.Abs(target)
	if err != nil {
		return "", false
	}

	if !strings.HasPrefix(target, base+string(filepath.Separator)) && target != base {
		return "", false
	}

	return target, true
}

func RandomString(length int) string {
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)

	for i := range b {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		b[i] = charset[n.Int64()]
	}

	return string(b)
}

func DetectContentType(reader io.Reader) (mime string, sample []byte, err error) {
	sample = make([]byte, 512)

	n, err := reader.Read(sample)
	if err == io.EOF {
		err = nil
	}

	if err != nil {
		sample = nil
		return
	}
	sample = sample[:n]

	mime = http.DetectContentType(sample)

	return
}

func IsDevelopment() bool {
	appEnv := os.Getenv("APP_ENV")
	return strings.EqualFold(appEnv, "development")
}

func CopyMap[K comparable, V any](m map[K]V) map[K]V {
	res := make(map[K]V, len(m))

	for key, val := range m {
		res[key] = val
	}

	return res
}

// 如果 val 不是nil 则会将 val 赋值给target
func SetIfNotNil[T any](target *T, val *T) {
	if val != nil {
		*target = *val
	}
}

// 如果要执行SetIfNotNil的项太多可以使用这个函数来简化操作
func MultiSetIfNotNil[T any](pairs [][]*T) {
	for _, item := range pairs {
		SetIfNotNil(item[0], item[1])
	}
}

func IsEmptyString(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}
