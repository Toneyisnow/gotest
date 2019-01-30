package log// +build debug

import (
log "github.com/sirupsen/logrus"
)

var (
	log_level = log.InfoLevel
)

// 日志，级别：调试。只在测试环境输出
func D(v ...interface{}) {
	// 只在测试环境生效
	if _log != nil {
		// debug.Println(v...)
		// debug.Output(LOG_DEPTH, fmt.Sprintln(v...))
		_log.Debugln(v...)
	}
}

// 日志，级别：调试。只在测试环境输出
func D2(format string, v ...interface{}) {
	// 只在测试环境生效
	if _log != nil {
		// debug.Println(v...)
		// debug.Output(LOG_DEPTH, fmt.Sprintf(format, v...))
		_log.Debugf(format, v...)
	}
}

