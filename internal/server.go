package internal

import (
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/routes"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func CreateServer(address string) {
	initializer.LoadAppConfig()            // load ./configs/config.yaml
	err := initializer.GormSQLServerInit() // open connection to sql server with GORM
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery())

	routes.LoadRoutes(engine)

	engine.GET("/image", func(ctx *gin.Context) {
		file, err := os.Open("kotamalang.jpg")

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Invalid Image",
			})
		}

		defer file.Close()
		fileInfo, _ := file.Stat()
		fileSize := fileInfo.Size()

		ctx.DataFromReader(http.StatusOK, fileSize, "image/jpg", file, nil)
	})

	engine.Run(address)
}