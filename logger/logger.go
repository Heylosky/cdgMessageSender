package logger

import (
	"github.com/messagebird/go-rest-api/v9/sms"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
	"os"
)

type LogConf struct {
	filename   string
	mode       string
	level      zapcore.Level
	maxSize    int
	maxAge     int
	maxBackups int
}

type Option func(*LogConf)

func Mode(m string) Option {
	return func(lc *LogConf) {
		lc.mode = m
	}
}

func Level(l zapcore.Level) Option {
	return func(lc *LogConf) {
		lc.level = l
	}
}

func InitLogger(filename string, opts ...Option) (err error) {
	//读取配置，没有就用默认
	lc := &LogConf{
		mode:       "dev",              //模式
		filename:   filename,           //日志存放路径
		level:      zapcore.DebugLevel, //日志级别
		maxSize:    200,                //最大存储大小，MB
		maxAge:     30,                 //最大存储时间
		maxBackups: 10,                 //备份数量
	}
	for _, opt := range opts {
		opt(lc)
	}

	//创建核心三大件，进行初始化
	//NewCore(enc Encoder, ws WriteSyncer, enab LevelEnabler)
	writerSyncer := getLogWriter(filename, lc.maxSize, lc.maxBackups, lc.maxAge)
	encoder := getEncoder()

	//创建核心
	var core zapcore.Core
	//如果是dev模式，同时要在前端打印；如果是其他模式，就只输出到文件
	if lc.mode == "dev" {
		//使用默认的encoder配置就行了
		//NewConsoleEncoder里面实际上就是一个NewJSONEncoder，需要输入配置
		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		//Tee方法将全部日志条目复制到两个或多个底层核心中
		core = zapcore.NewTee(
			zapcore.NewCore(encoder, writerSyncer, lc.level),                   //写入到文件的核心
			zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), lc.level), //写到前台的核心
		)
	} else {
		core = zapcore.NewCore(encoder, writerSyncer, lc.level)
	}

	//创建logger对象
	//New方法返回logger，非自定义的情况下就是NewProduction, NewDevelopment,NewExample或者config就可以了。
	//zap.AddCaller是个option，会添加上调用者的文件名和行数，到日志里
	logger := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(logger)

	// return logger 如果return了logger就可以使用之前的ginzap.Ginzap和ginzap.RecoveryWithZap。
	return
}

func getLogWriter(filename string, maxSize, maxBackup, maxAge int) zapcore.WriteSyncer {
	//使用lumberjack分割归档日志
	lumberLackLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxAge:     maxAge,
		MaxBackups: maxBackup,
	}
	return zapcore.AddSync(lumberLackLogger)
}

func getEncoder() zapcore.Encoder {
	//使用一份官方预定义的production的配置，然后更改
	encoderConfig := zap.NewProductionEncoderConfig()
	//默认时间格式是这样的: "ts":1670214777.9225469 | EpochTimeEncoder serializes a time.Time to a floating-point number of seconds
	//重新设置时间格式
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	//重新设置时间字段的key
	encoderConfig.TimeKey = "time"
	//默认的level是小写的zapcore.LowercaseLevelEncoder ｜ "level":"info" 可以改成大写
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	//"caller":"zap/zap.go:90" 也可以改成Full的更加详细
	//encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

func SMSLogger(message *sms.Message) {
	zap.L().Info("send a message to MB",
		zap.String("Originator", message.Originator),
		zap.String("Body", message.Body),
		zap.Any("Recipients", message.Recipients),
		zap.Timep("Timestamp", message.CreatedDatetime),
	)
}

func SMSLogOnError(originator string, body string, recipient string) {
	zap.L().Error("Message sending failed. First time retry failed.",
		zap.String("Originator", originator),
		zap.String("Body", body),
		zap.String("Recipients", recipient),
	)
}
