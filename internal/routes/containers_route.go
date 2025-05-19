package routes

import (
	"future-interns-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func ContainersRoute(apiv1 *gin.RouterGroup) {
	handlers := &handlers.ContainersHandler{}

	router := apiv1.Group("/containers")
	{
		router.Handle("GET", "/", handlers.ListContainers)
		router.Handle("POST", "/", handlers.ContainersControl)
	}
}
