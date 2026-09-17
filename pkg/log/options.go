package log

import "io"

// Options 描述全局 logger 的构建参数。
type Options struct {
	// Level 为日志级别（debug|info|warn|error），空值回退为 info。
	Level string
	// Console 表示是否输出到标准输出。
	Console bool
	// Writers 为额外的输出目标（例如 lumberjack 文件轮换）。nil 元素会被忽略。
	Writers []io.Writer
}
