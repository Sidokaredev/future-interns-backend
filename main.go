package main

import (
	initializer "go-write-behind-service/init"
	"go-write-behind-service/internal/middlewares"
	"go-write-behind-service/internal/routes"
	"go-write-behind-service/internal/scheduler"
	"time"

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

	go scheduler.StartWriterJob(2 * time.Minute)
	go scheduler.StartUpdaterJob(4 * time.Minute)

	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), middlewares.CORSPolicy())

	routes.ListenToRoutes(engine)

	engine.Run(":8003")
}
