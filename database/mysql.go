package database

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func NewMysqlClient(opt *MysqlOptions) (*gorm.DB, error) {
	if opt == nil {
		return nil, errors.New("option is nil")
	}
	if opt.Source == nil || len(opt.Source.Infos) == 0 {
		return nil, errors.New("source is nil")
	}

	// init config
	config := &gorm.Config{
		SkipDefaultTransaction: true,
	}
	if opt.Log != nil {
		config.Logger = NewLog(opt.Log, opt.Debug)
	}

	var (
		sources  []gorm.Dialector
		replicas []gorm.Dialector
		resolver = &dbresolver.DBResolver{}
	)
	// sources
	for _, source := range opt.Source.Infos {
		sources = append(sources, getMysqlDialector(source))
	}
	resolver = resolver.Register(
		dbresolver.Config{
			Sources:           sources,
			Policy:            dbresolver.RandomPolicy{},
			TraceResolverMode: true,
		},
		opt.Source.Objs...,
	)
	// replicas
	if opt.Replicas != nil {
		for _, replica := range opt.Replicas.Infos {
			replicas = append(replicas, getMysqlDialector(replica))
		}
		resolver = resolver.Register(
			dbresolver.Config{
				Replicas:          replicas,
				Policy:            dbresolver.RandomPolicy{},
				TraceResolverMode: true,
			},
			opt.Replicas.Objs...,
		)
	}

	// open db
	d, err := gorm.Open(sources[0], config)
	if err != nil {
		return nil, err
	}
	// register resolver plugin
	if err = d.Use(resolver); err != nil {
		return nil, err
	}

	db, err := d.DB()
	if err != nil {
		return nil, err
	}

	// set connection configuration
	if opt.MaxIdleTime != "" {
		if maxIdleTime, err := time.ParseDuration(opt.MaxIdleTime); err != nil {
			return nil, err
		} else {
			db.SetConnMaxIdleTime(maxIdleTime)
		}
	}
	if opt.MaxIdleConn != 0 {
		db.SetMaxIdleConns(opt.MaxIdleConn)
	}
	if opt.MaxOpenConn != 0 {
		db.SetMaxOpenConns(opt.MaxOpenConn)
	}

	// auto migration
	if opt.Source.Migrate {
		err = d.Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4").
			Clauses(dbresolver.Write).
			Migrator().
			AutoMigrate(opt.Source.Objs...)
		if err != nil {
			return nil, err
		}
	}

	return d, nil
}

func getMysqlDialector(info *BackendInfo) gorm.Dialector {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", info.User, info.Passwd,
		info.Host, info.Port, info.DB)
	return mysql.New(mysql.Config{
		DSN:                       dsn,
		DriverName:                "mysql",
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		SkipInitializeWithVersion: false,
	})
}
