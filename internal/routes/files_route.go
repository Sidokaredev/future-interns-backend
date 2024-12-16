package routes

import (
	"future-interns-backend/internal/constants"
	"future-interns-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func ImageRoutes(apiv1 *gin.RouterGroup) {
	fileHandlers := &handlers.FilesHandlers{}

	router := apiv1.Group("/images")
	{
		router.Handle(constants.MethodGet, "/:id", fileHandlers.GetById)
		router.Handle(constants.MethodPost, "/", fileHandlers.Create)
	}
	routerDocument := apiv1.Group("/documents")
	{
		routerDocument.Handle(constants.MethodGet, "/:id", fileHandlers.DocumentGetById)
		routerDocument.Handle(constants.MethodGet, "/:id/download", fileHandlers.DocumentDownloadById)
	}
}
