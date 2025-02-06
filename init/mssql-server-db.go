package initializer

import (
	"errors"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

var (
	mssqlDB *gorm.DB
)

func MssqlServerInit() error {
	var dbconfig struct {
		Username          string
		Password          string
		Host              string
		Port              int
		Instance          string
		Database          string
		ConnectionTimeout int
		Encrypt           bool
	}
	if err := viper.UnmarshalKey("mssql.dev", &dbconfig); err != nil {
		return err
	}

	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", dbconfig.Username, dbconfig.Password, dbconfig.Host, dbconfig.Port, dbconfig.Database)

	var errDB error
	mssqlDB, errDB = gorm.Open(sqlserver.Open(dsn), &gorm.Config{TranslateError: true})
	if errDB != nil {
		return errDB
	}

	log.Println("mssql server connection established")
	return nil
}

func GetMssqlDB() (*gorm.DB, error) {
	if mssqlDB == nil {
		return nil, errors.New("mssql server database instance empty")
	}

	return mssqlDB, nil
}
