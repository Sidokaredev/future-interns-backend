package initializer

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/microsoft/go-mssqldb"
	"github.com/spf13/viper"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var (
	sqldb     *sql.DB
	gormsqldb *gorm.DB
	once      sync.Once
)
var (
	dbconfig DatabaseProps
	errDB    error
)

func MssqlInit() error {
	if err := viper.UnmarshalKey("databases.dev", &dbconfig); err != nil {
		panic(err)
	}

	// var errDB error
	once.Do(func() {
		dataSource := fmt.Sprintf("sqlserver://%s:%s@%s?database=%s",
			dbconfig.Username, dbconfig.Password, dbconfig.Host, dbconfig.Database)

		sqldb, errDB = sql.Open("sqlserver", dataSource)
		if errDB != nil {
			errDB = fmt.Errorf("failed to open connection to sql server database: %s", errDB)
			return
		}

		sqldb.SetMaxOpenConns(10)
		sqldb.SetMaxIdleConns(5)
		// sqldb.SetConnMaxLifetime(30 * time.Minute)

		if errDB = sqldb.Ping(); errDB != nil {
			errDB = fmt.Errorf("test connection failes : %w", errDB)
			return
		}

		log.Println("sql server connection established...")
	})

	return errDB
}

func GormSQLServerInit() error {
	if err := viper.UnmarshalKey("databases.dev", &dbconfig); err != nil {
		panic(err)
	}

	once.Do(func() {
		dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", dbconfig.Username, dbconfig.Password, dbconfig.Host, dbconfig.Port, dbconfig.Database)

		gormsqldb, errDB = gorm.Open(sqlserver.Open(dsn), &gorm.Config{})
		if errDB != nil {
			errDB = fmt.Errorf("failed to open connection to sql server database using GORM \t: %w", errDB)
			return
		}

		if sqldb, errDB = gormsqldb.DB(); errDB != nil {
			errDB = fmt.Errorf("failed getting sql.DB instance \t: %w", errDB)
			return
		}

		sqldb.SetMaxOpenConns(10)
		sqldb.SetMaxIdleConns(5)
		sqldb.SetConnMaxIdleTime(30 * time.Minute)

		if errDB = sqldb.Ping(); errDB != nil {
			errDB = fmt.Errorf("test connection failes : %w", errDB)
			return
		}

		log.Println("sql server connection established with GORM...")
	})

	return errDB
}

func GetGorm() (*gorm.DB, error) {
	if gormsqldb == nil {
		return nil, fmt.Errorf("database cannot be initialized - GORM")
	}
	return gormsqldb, nil
}

func CallConn() (*sql.DB, error) {
	if sqldb == nil {
		return nil, fmt.Errorf("database cannot be initialized")
	}
	return sqldb, nil
}
