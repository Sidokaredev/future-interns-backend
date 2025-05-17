package initializer

import (
	"fmt"
	"net/url"

	"github.com/spf13/viper"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

type DatabaseConfigProps struct {
	Username          string
	Password          string
	Host              string
	Port              int
	Database          string
	ConnectionTimeout int
	Encrypt           bool
}

var (
	gormsqldb *gorm.DB
	dbconfig  DatabaseConfigProps
)

func GormSQLServer() {
	if errConfig := viper.UnmarshalKey("databases.testing", &dbconfig); errConfig != nil {
		panic(errConfig)
	}

	DataSourceName := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", dbconfig.Username, url.QueryEscape(dbconfig.Password), dbconfig.Host, dbconfig.Port, dbconfig.Database)

	gormDB, errGorm := gorm.Open(sqlserver.Open(DataSourceName), &gorm.Config{TranslateError: true})
	if errGorm != nil {
		panic(errGorm)
	}

	gormsqldb = gormDB
	fmt.Println("GORM \t:", " sql server connection established")
}

func GetGorm() (*gorm.DB, error) {
	if gormsqldb == nil {
		return nil, fmt.Errorf("GORM : sql server connection not established")
	}
	return gormsqldb, nil
}
