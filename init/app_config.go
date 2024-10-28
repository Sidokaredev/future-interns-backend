package initializer

import (
	"fmt"

	"github.com/spf13/viper"
)

type Configs struct {
	Databases struct {
		Dev  DatabaseProps
		Prod DatabaseProps
	}
}

type DatabaseProps struct {
	Username          string
	Password          string
	Host              string
	Port              int
	Instance          string
	Database          string
	ConnectionTimeout int
	Encrypt           bool
}

func LoadAppConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
		// return fmt.Errorf("error while read in config ./configs/config.yaml: %w", err)
	}

	fmt.Println("the config has been configured using Viper...")
}
