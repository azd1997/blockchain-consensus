/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/19/20 9:10 AM
* @Description: The file is for
***********************************************************************/

package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

// LogDest 日志目的地
type LogDest string

const (
	LogDest_Stderr LogDest = "stderr"
	LogDest_File   LogDest = "file"
)

// Logger 简单的日志器，自定义输出路径
type Logger struct {
	logger *log.Logger
	module string
	id     string
	dest   LogDest
}

// NewLogger
func NewLogger(dest LogDest, module string, id string) *Logger {
	prefix := fmt.Sprintf("[%s](%s) ", module, id)
	var w io.Writer
	switch dest {
	case LogDest_Stderr:
		w = os.Stderr
	case LogDest_File:
		filename := fmt.Sprintf("./%s-%s.log", module, id)
		f, err := os.Create(filename)
		if err != nil {
			log.Panicf("create log file fail: %s\n", err)
		}
		w = f
	default:
		log.Panic("unknown log destination")
	}

	return &Logger{
		logger: log.New(w, prefix, log.LstdFlags),
		module: module,
		id:     id,
		dest:   dest,
	}
}

// Log 打日志
func (logger *Logger) Log(v ...interface{}) {
	logger.logger.Println("[L-O-G] ", v)
}

// Logf 打日志
func (logger *Logger) Logf(format string, v ...interface{}) {
	format = fmt.Sprintf("[L-O-G] %s", format)
	logger.logger.Printf(format, v...)
}

// Error 报错
func (logger *Logger) Error(v ...interface{}) {
	logger.logger.Println("[ERROR] ", v)
}

// Errorf 报错
func (logger *Logger) Errorf(format string, v ...interface{}) {
	format = fmt.Sprintf("[ERROR] %s", format)
	logger.logger.Printf(format, v...)
}

// Fatal 致命错误
func (logger *Logger) Fatal(v ...interface{}) {
	logger.logger.Fatalln("[FATAL] ", v)
}

// Fatalf 致命错误
func (logger *Logger) Fatalf(format string, v ...interface{}) {
	format = fmt.Sprintf("[FATAL] %s", format)
	logger.logger.Fatalf(format, v...)
}

// Panic 恐慌
func (logger *Logger) Panic(v ...interface{}) {
	logger.logger.Panicln("[PANIC] ", v)
}

// Panicf 恐慌
func (logger *Logger) Panicf(format string, v ...interface{}) {
	format = fmt.Sprintf("[PANIC] %s", format)
	logger.logger.Panicf(format, v...)
}
