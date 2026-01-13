package database

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Log struct {
	isDebug bool
	logger  *logrus.Logger
}

func (l *Log) Info(ctx context.Context, s string, i ...interface{}) {
	l.logger.Infof(s, i...)
}

func (l *Log) Warn(ctx context.Context, s string, i ...interface{}) {
	l.logger.Warnf(s, i...)
}

func (l *Log) Error(ctx context.Context, s string, i ...interface{}) {
	l.logger.Errorf(s, i...)
}

func (l *Log) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elasped := time.Since(begin)
	sql, rows := fc()
	if err != nil {
		// record not found 类型不进行记录
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return
		}
		l.logger.Errorf("trace: %s, [rows: %d] %s, error: %v", elasped, rows, sql, err)
	} else {
		if l.isDebug {
			l.logger.Debugf("trace: %s, [rows: %d] %s", elasped, rows, sql)
		}
	}
}

func (l *Log) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	switch level {
	case logger.Silent:
		newLogger.logger.SetLevel(logrus.PanicLevel)
	case logger.Error:
		newLogger.logger.SetLevel(logrus.ErrorLevel)
	case logger.Warn:
		newLogger.logger.SetLevel(logrus.WarnLevel)
	case logger.Info:
		newLogger.logger.SetLevel(logrus.DebugLevel)
	default:
		newLogger.logger.SetLevel(logrus.DebugLevel)
	}
	return &newLogger
}

func NewLog(logger *logrus.Logger, debug bool) *Log {
	return &Log{logger: logger, isDebug: debug}
}
