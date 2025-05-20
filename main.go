package main

import (
	initializer "go-no-cache-service/init"
	"go-no-cache-service/internal/middlewares"
	"go-no-cache-service/internal/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	if errLoadConfig := initializer.LoadAppConfig(); errLoadConfig != nil {
		panic(errLoadConfig)
	}

	if errMssql := initializer.MssqlServerInit(); errMssql != nil {
		panic(errMssql)
	}

	if errRedis := initializer.RedisServerInit(); errRedis != nil {
		panic(errRedis)
	}

	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), middlewares.CORSPolicy())

	routes.ListenToRoutes(engine)
	engine.Run(":8004")
}
