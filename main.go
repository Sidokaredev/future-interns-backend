package main

import (
	initializer "go-cache-aside-service/init"
	"go-cache-aside-service/internal/middlewares"
	"go-cache-aside-service/internal/routes"
	"log"

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
		log.Println("redis :" + errRedis.Error())
		panic(errRedis)
	}

	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), middlewares.CORSPolicy())

	routes.ListenToRoutes(engine)
	engine.Run(":8000")
}
