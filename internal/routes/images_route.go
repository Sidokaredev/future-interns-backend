package routes

import (
	"future-interns-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func ImageRoutes(apiv1 *gin.RouterGroup) {
	imageHandlers := &handlers.ImageHandlers{}

	router := apiv1.Group("/images")
	{
		router.Handle(MethodGet, "/:id", imageHandlers.GetById)
	}
}
