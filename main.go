package main

import (
	initializer "go-read-through-service/init"
	"go-read-through-service/internal/middlewares"
	"go-read-through-service/internal/routes"

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

	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), middlewares.CORSPolicy())

	routes.ListenToRoutes(engine)
	engine.Run(":8000")
}
