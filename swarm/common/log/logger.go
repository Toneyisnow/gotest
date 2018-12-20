package log

import (
	"os"
	log "github.com/sirupsen/logrus"
)

const (
	LOG_DEPTH int = 2
)

var (
	// info  *log.Logger
	// warn  *log.Logger
	// err   *log.Logger
	// debug *log.Logger

	_log *log.Logger
)

// 初始化日志
func init() {
	Log2File(os.Stdout)
}

func Log2File(outFile *os.File) {
	// flag := log.LstdFlags | log.Lmicroseconds
	// info = log.New(outFile, "[INFO] ", flag|log.Lshortfile)
	// warn = log.New(outFile, "[WARN] ", flag)
	// err = log.New(outFile, "[ERROR] ", flag|log.Lshortfile)
	// debug = log.New(outFile, "[DEBUG] ", flag|log.Lshortfile)
	_log = log.New()
	_log.SetOutput(outFile)
	_log.SetLevel(log_level)
}

// 日志，级别：信息
func I(v ...interface{}) {
	if _log != nil {
		_log.Infoln(v...)
	}
}

// 日志，级别：信息
func I2(format string, v ...interface{}) {
	if _log != nil {
		_log.Infof(format, v...)
	}
}

// 日志，级别：警告
func W(v ...interface{}) {
	if _log != nil {
		_log.Warnln(v...)
	}
}

// 日志，级别：警告
func W2(format string, v ...interface{}) {
	if _log != nil {
		_log.Warnf(format, v...)
	}
}

// 日志，级别：错误。输出并退出程序
func E(v ...interface{}) {
	if _log != nil {
		// err.Fatalln(v...)
		_log.Panicln(v...)
		//os.Exit(1)
	}
}

// 日志，级别：错误。输出并退出程序
func E2(format string, v ...interface{}) {
	if _log != nil {
		// err.Fatalln(v...)
		_log.Panicf(format, v...)
		//os.Exit(1)
	}
}

// 日志，级别：信息。带有字段信息
func F(fields map[string]interface{}, v ...interface{}) {
	_log.WithFields(fields).Infoln(v)
}


func Logger() *log.Logger {
	return _log
}

