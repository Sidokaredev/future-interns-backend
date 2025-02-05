package main

import (
	initializer "go-cache-aside-service/init"
	"go-cache-aside-service/internal/middlewares"
	"go-cache-aside-service/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	if errLoadConfig := initializer.LoadAppConfig(); errLoadConfig != nil {
		panic(errLoadConfig)
	}

	if errMssql := initializer.MssqlServerInit(); errMssql != nil {
		panic(errMssql)
	}

	// gormDB, errGorm := initializer.GetMssqlDB()
	// if errGorm != nil {
	// 	log.Fatal(errGorm)
	// }

	// gormDB.AutoMigrate(&models.CacheSession{}, &models.RequestLog{})

	if errRedis := initializer.RedisServerInit(); errRedis != nil {
		panic(errRedis)
	}

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), middlewares.CORSPolicy())

	routes.ListenToRoutes(engine)
	engine.Run(":8000")
}
