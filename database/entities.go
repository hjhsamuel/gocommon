package database

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type BackendConfig struct {
	Debug       bool           `yaml:"debug" env:"DEBUG"`               // 是否开启debug模式
	MaxIdleConn int            `yaml:"max_idle_conn" env:"MAXIDLECONN"` // 最大空闲连接数
	MaxOpenConn int            `yaml:"max_open_conn" env:"MAXOPENCONN"` // 最大连接数
	Sources     *SourceInfo    `yaml:"sources" env:"SOURCES"`           // source 配置
	Replicas    []*BackendInfo `yaml:"replicas" env:"REPLICAS"`         // replica 配置
}

type SourceInfo struct {
	Migrate bool           `yaml:"migrate" env:"MIGRATE"`
	Infos   []*BackendInfo `yaml:"infos" env:"INFOS"`
}

type BackendInfo struct {
	Host   string `yaml:"host" env:"HOST"`     // 地址
	Port   int    `yaml:"port" env:"PORT"`     // 端口
	DB     string `yaml:"db" env:"DB"`         // 数据库名称
	User   string `yaml:"user" env:"USER"`     // 账号
	Passwd string `yaml:"passwd" env:"PASSWD"` // 密码
}

func (c *BackendInfo) UnmarshalText(text []byte) error {
	reflectMap := make(map[string]string)
	t := reflect.TypeOf(*c)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tagName := field.Tag.Get("env")
		reflectMap[tagName] = field.Name
	}

	v := reflect.ValueOf(c).Elem()

	parts := strings.Split(string(text), "|")
	for _, part := range parts {
		index := strings.Index(part, "=")
		if index == -1 {
			return fmt.Errorf("invalid backend info: %s", part)
		}
		if name, ok := reflectMap[part[:index]]; ok {
			if field, rok := t.FieldByName(name); rok {
				switch field.Type.Kind() {
				case reflect.String:
					v.FieldByName(name).SetString(part[index+1:])
				case reflect.Int:
					intVal, err := strconv.Atoi(part[index+1:])
					if err != nil {
						return err
					}
					v.FieldByName(name).SetInt(int64(intVal))
				default:
					continue
				}
			}
		}
	}
	return nil
}

type MysqlOptions struct {
	Log         *logrus.Logger
	Debug       bool
	MaxIdleConn int
	MaxOpenConn int
	MaxIdleTime string
	Source      *MysqlDialector
	Replicas    *MysqlDialector
}

type MysqlDialector struct {
	SourceInfo
	Objs []any
}

type MongoConfig struct {
	Debug        bool              `yaml:"debug" env:"DEBUG"`
	MaxIdleConns uint64            `yaml:"max_idle_conn" env:"MAXIDLECONN"`
	MinIdleConns uint64            `yaml:"min_idle_conn" env:"MINIDLECONN"`
	MaxIdleTime  string            `yaml:"max_idle_time" env:"MAXIDLETIME"`
	User         string            `yaml:"user" env:"USER"`
	Passwd       string            `yaml:"passwd" env:"PASSWD"`
	Addr         []*MongoAddr      `yaml:"addr" env:"ADDR"`
	DB           string            `yaml:"db" env:"DB"`
	Opts         map[string]string `yaml:"opts" env:"OPTS"`
}

type MongoAddr struct {
	Host string `yaml:"host" env:"HOST"`
	Port int    `yaml:"port" env:"PORT"`
}

func (m *MongoAddr) UnmarshalText(text []byte) error {
	reflectMap := make(map[string]string)
	t := reflect.TypeOf(*m)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tagName := field.Tag.Get("env")
		reflectMap[tagName] = field.Name
	}

	v := reflect.ValueOf(m).Elem()

	parts := strings.Split(string(text), "|")
	for _, part := range parts {
		index := strings.Index(part, "=")
		if index == -1 {
			return fmt.Errorf("invalid MongoAddr: %s", part)
		}
		if name, ok := reflectMap[part[:index]]; ok {
			if field, rok := t.FieldByName(name); rok {
				switch field.Type.Kind() {
				case reflect.String:
					v.FieldByName(name).SetString(part[index+1:])
				case reflect.Int:
					intVal, err := strconv.Atoi(part[index+1:])
					if err != nil {
						return err
					}
					v.FieldByName(name).SetInt(int64(intVal))
				}
			}
		}
	}
	return nil
}
