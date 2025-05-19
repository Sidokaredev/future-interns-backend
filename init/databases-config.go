package initializer

import (
	"log"

	"github.com/spf13/viper"
)

func LoadAppConfig() error {
	viper.SetConfigName("databases")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	log.Println("databases.yaml loaded ...")

	return nil
}
