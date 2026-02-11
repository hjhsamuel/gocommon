package database

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bombsimon/logrusr/v4"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewMongoClient(opt *MongoConfig, logger *logrus.Logger) (*mongo.Client, error) {
	dsn, err := getMongDialector(opt)
	if err != nil {
		return nil, err
	}

	mongoOpts := options.Client().ApplyURI(dsn)
	if opt.MaxIdleConns != 0 {
		mongoOpts.SetMaxPoolSize(opt.MaxIdleConns)
	}
	if opt.MinIdleConns != 0 {
		mongoOpts.SetMinPoolSize(opt.MinIdleConns)
	}
	if opt.MaxIdleTime != "" {
		du, err := time.ParseDuration(opt.MaxIdleTime)
		if err != nil {
			return nil, err
		}
		mongoOpts.SetMaxConnIdleTime(du)
	}

	if logger != nil {
		sink := logrusr.New(logger).GetSink()
		loggerOpts := options.Logger().SetSink(sink)

		var logLevel options.LogLevel
		if opt.Debug {
			logLevel = options.LogLevelDebug
		} else {
			logLevel = options.LogLevelInfo
		}
		loggerOpts.SetComponentLevel(options.LogComponentCommand, logLevel)

		mongoOpts.SetLoggerOptions(loggerOpts)
	}

	client, err := mongo.Connect(mongoOpts)
	if err != nil {
		return nil, err
	}

	return client, nil
}

func getMongDialector(info *MongoConfig) (string, error) {
	parts := make([]string, 0)
	parts = append(parts, "mongodb://")

	// set credentials
	if info.User != "" {
		parts = append(parts, info.User)
		if info.Passwd != "" {
			parts = append(parts, ":"+info.Passwd)
		}
		parts = append(parts, "@")
	}
	// set addresses
	addrs := make([]string, 0)
	for _, item := range info.Addr {
		if item.Port != 0 {
			addrs = append(addrs, fmt.Sprintf("%s:%d", item.Host, item.Port))
		} else {
			addrs = append(addrs, item.Host)
		}
	}
	if len(addrs) == 0 {
		return "", errors.New("no address found")
	}
	parts = append(parts, strings.Join(addrs, ","), "/")
	// set default database
	if info.DB != "" {
		parts = append(parts, info.DB)
	}
	// set options
	clientOptions := make([]string, 0)
	for k, v := range info.Opts {
		clientOptions = append(clientOptions, fmt.Sprintf("%s=%s", k, v))
	}
	if len(clientOptions) != 0 {
		parts = append(parts, "?", strings.Join(clientOptions, ";"))
	}

	return strings.Join(parts, ""), nil
}
