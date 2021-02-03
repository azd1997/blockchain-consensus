/**********************************************************************
* @Author: Eiger (201820114847@mail.scut.edu.cn)
* @Date: 12/14/20 6:11 PM
* @Description: 基于zap的log
***********************************************************************/

package log

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

// 全局单例
// 一般情况下都是使用全局单例，但是对于想要在内存中进行集群部署来说，需要有logger生成函数
var loggers = map[string]*zap.SugaredLogger{} // <id, logger>

// 独立设出来，是为了方便在程序中关闭所使用的日志文件
var lumberjackes = map[string]*lumberjack.Logger{}

//var (
//	highPriority = zap.LevelEnablerFunc(func(level zapcore.Level) bool {
//		return level >= zapcore.InfoLevel
//	})
//
//	lowPriority = zap.LevelEnablerFunc(func(level zapcore.Level) bool {
//		return level >= zapcore.DebugLevel
//	})
//)

// InitGlobalLogger 初始化全局日志单例
func InitGlobalLogger(id string, debug bool, addCaller bool, logFileName ...string) {
	if loggers[id] != nil {
		return
	}
	if len(logFileName) > 1 {
		panic("too much log file specified")
	}

	var allCore []zapcore.Core

	var logLevel zapcore.LevelEnabler
	if debug {
		logLevel = zapcore.DebugLevel
	} else {
		logLevel = zapcore.InfoLevel
	}

	encoder := getEncoder()
	consoleWriter := zapcore.Lock(os.Stdout)
	consoleCore := zapcore.NewCore(encoder, consoleWriter, logLevel)

	if len(logFileName) > 0 { // 只取第1个文件地址

		lj := &lumberjack.Logger{
			Filename:   logFileName[0],
			MaxSize:    1,
			MaxAge:     30,
			MaxBackups: 5,
			LocalTime:  false,
			Compress:   false,
		}
		lumberjackes[id] = lj

		fileWriter := zapcore.AddSync(lj)
		fileCore := zapcore.NewCore(encoder, fileWriter, logLevel)
		allCore = append(allCore, fileCore)
	}

	allCore = append(allCore, consoleCore)

	core := zapcore.NewTee(allCore...)

	var logger *zap.Logger
	if addCaller {
		// 由于我们不直接使用sugarLogger，而是再包装一次，所以caller要再skip一次
		logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	} else {
		logger = zap.New(core)
	}

	loggers[id] = logger.Sugar()
}

// Close 关闭对应logger所持有的文件
func Close(id string) error {
	if loggers[id] != nil {
		loggers[id].Sync() // 这里的错误需要忽略
		if lumberjackes[id] != nil {
			if err := lumberjackes[id].Close(); err != nil {
				return err
			}
		}
	}
	return nil
}

// InitGlobalLogger 初始化全局日志单例
//func InitGlobalLogger(logFileName ...string) {
//	if len(logFileName) > 1 {panic("too much log file specified")}
//
//	var allCore []zapcore.Core
//
//	encoder := getEncoder()
//	consoleWriter := zapcore.Lock(os.Stdout)
//	consoleCore := zapcore.NewCore(encoder, consoleWriter, zapcore.DebugLevel)
//
//
//	if len(logFileName) > 0 {	// 只取第1个文件地址
//		fileWriter := getLogWriter(logFileName[0])
//		fileCore := zapcore.NewCore(encoder, fileWriter, zapcore.DebugLevel)
//		allCore = append(allCore, fileCore)
//	}
//
//	allCore = append(allCore, consoleCore)
//
//	core := zapcore.NewTee(allCore...)
//
//	// 由于我们不直接使用sugarLogger，而是再包装一次，所以caller要再skip一次
//	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
//	sugarLogger = logger.Sugar()
//}

// Sync 将日志写到文件中去
// 需要注意，由于这里使用的zaplogger有两个输出，一个是控制台，一个是文件。 Sync对Stdout是无效的，会报错，所以只能忽略这个错误了
func Sync(id string) {
	//if err := sugarLogger.Sync(); err != nil {
	//	panic(err)
	//}
	//sugarLogger.Sync()

	if loggers[id] != nil {
		loggers[id].Sync()

	}
	//for id := range loggers {
	//	loggers[id].Sync()
	//}
}

func getLogWriter(filename string) zapcore.WriteSyncer {
	lumberjackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    1,
		MaxAge:     30,
		MaxBackups: 5,
		LocalTime:  false,
		Compress:   false,
	}
	return zapcore.AddSync(lumberjackLogger)
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}
