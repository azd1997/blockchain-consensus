/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 10/19/20 9:10 AM
* @Description: The file is for
***********************************************************************/

package log

import (
	"fmt"
)

// Logger 基于sugarLogger包装
type Logger struct {
	module string
	id     string
	prefix string
}

// NewLogger 新建
func NewLogger(module string, id string) *Logger {
	if loggers[id] == nil {
		InitGlobalLogger(id, true, true)
		fmt.Printf("loggers[id] == nil。 create a logger {id=%s, debug=%v, addCaller=%v, logFileName=%v}\n", id, true, true, nil)
	}

	prefix := fmt.Sprintf("\t[%s](%s)\t", module, id)

	return &Logger{
		module: module,
		id:     id,
		prefix: prefix,
	}
}

func (logger *Logger) Debug(args ...interface{}) {
	args = append([]interface{}{logger.prefix}, args...)
	loggers[logger.id].Debug(args...)
}

func (logger *Logger) Debugf(template string, args ...interface{}) {
	template = fmt.Sprintf("%s%s", logger.prefix, template)
	loggers[logger.id].Debugf(template, args...)
}

func (logger *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	msg = fmt.Sprintf("%s%s", logger.prefix, msg)
	loggers[logger.id].Debugw(msg, keysAndValues...)
}

func (logger *Logger) Info(args ...interface{}) {
	args = append([]interface{}{logger.prefix}, args...)
	loggers[logger.id].Info(args...)
}

func (logger *Logger) Infof(template string, args ...interface{}) {
	template = fmt.Sprintf("%s%s", logger.prefix, template)
	loggers[logger.id].Infof(template, args...)
}

func (logger *Logger) Infow(msg string, keysAndValues ...interface{}) {
	msg = fmt.Sprintf("%s%s", logger.prefix, msg)
	loggers[logger.id].Infow(msg, keysAndValues...)
}

func (logger *Logger) Warn(args ...interface{}) {
	args = append([]interface{}{logger.prefix}, args...)
	loggers[logger.id].Warn(args...)
}

func (logger *Logger) Warnf(template string, args ...interface{}) {
	template = fmt.Sprintf("%s%s", logger.prefix, template)
	loggers[logger.id].Warnf(template, args...)
}

func (logger *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	msg = fmt.Sprintf("%s%s", logger.prefix, msg)
	loggers[logger.id].Warnw(msg, keysAndValues...)
}

func (logger *Logger) Error(args ...interface{}) {
	args = append([]interface{}{logger.prefix}, args...)
	loggers[logger.id].Error(args...)
}

func (logger *Logger) Errorf(template string, args ...interface{}) {
	template = fmt.Sprintf("%s%s", logger.prefix, template)
	loggers[logger.id].Errorf(template, args...)
}

func (logger *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	msg = fmt.Sprintf("%s%s", logger.prefix, msg)
	loggers[logger.id].Errorw(msg, keysAndValues...)
}

func (logger *Logger) Panic(args ...interface{}) {
	args = append([]interface{}{logger.prefix}, args...)
	loggers[logger.id].Panic(args...)
}

func (logger *Logger) Panicf(template string, args ...interface{}) {
	template = fmt.Sprintf("%s%s", logger.prefix, template)
	loggers[logger.id].Panicf(template, args...)
}

func (logger *Logger) Panicw(msg string, keysAndValues ...interface{}) {
	msg = fmt.Sprintf("%s%s", logger.prefix, msg)
	loggers[logger.id].Panicw(msg, keysAndValues...)
}

func (logger *Logger) Fatal(args ...interface{}) {
	args = append([]interface{}{logger.prefix}, args...)
	loggers[logger.id].Fatal(args...)
}

func (logger *Logger) Fatalf(template string, args ...interface{}) {
	template = fmt.Sprintf("%s%s", logger.prefix, template)
	loggers[logger.id].Fatalf(template, args...)
}

func (logger *Logger) Fatalw(msg string, keysAndValues ...interface{}) {
	msg = fmt.Sprintf("%s%s", logger.prefix, msg)
	loggers[logger.id].Fatalw(msg, keysAndValues...)
}

//
//// Log 打日志
//func (logger *Logger) Log(v ...interface{}) {
//	logger.logger.Println("[L-O-G] ", v)
//}
//
//// Logf 打日志
//func (logger *Logger) Logf(format string, v ...interface{}) {
//	format = fmt.Sprintf("[L-O-G] %s", format)
//	logger.logger.Printf(format, v...)
//}
//
//// Error 报错
//func (logger *Logger) Error(v ...interface{}) {
//	logger.logger.Println("[ERROR] ", v)
//}
//
//// Errorf 报错
//func (logger *Logger) Errorf(format string, v ...interface{}) {
//	format = fmt.Sprintf("[ERROR] %s", format)
//	logger.logger.Printf(format, v...)
//}
//
//// Fatal 致命错误
//func (logger *Logger) Fatal(v ...interface{}) {
//	logger.logger.Fatalln("[FATAL] ", v)
//}
//
//// Fatalf 致命错误
//func (logger *Logger) Fatalf(format string, v ...interface{}) {
//	format = fmt.Sprintf("[FATAL] %s", format)
//	logger.logger.Fatalf(format, v...)
//}
//
//// Panic 恐慌
//func (logger *Logger) Panic(v ...interface{}) {
//	logger.logger.Panicln("[PANIC] ", v)
//}
//
//// Panicf 恐慌
//func (logger *Logger) Panicf(format string, v ...interface{}) {
//	format = fmt.Sprintf("[PANIC] %s", format)
//	logger.logger.Panicf(format, v...)
//}
