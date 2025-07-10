package main

import (
	initializer "go-write-behind-service/init"
	"go-write-behind-service/internal/middlewares"
	"go-write-behind-service/internal/routes"
	service_schedulers "go-write-behind-service/internal/services/scheduler"
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

	go service_schedulers.RunPipelinesScheduler(30 * time.Minute)

	// go scheduler.StartWriterJob(2 * time.Minute)  // -> job writer
	// go scheduler.StartUpdaterJob(4 * time.Minute) // -> job updater

	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), middlewares.CORSPolicy())

	routes.ListenToRoutes(engine)

	engine.Run(":8003")
}
