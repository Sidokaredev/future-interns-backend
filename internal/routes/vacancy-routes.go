package routes

import (
	"go-write-through-service/internal/handlers"
	handler_candidates "go-write-through-service/internal/handlers/candidates"
	"go-write-through-service/internal/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

func VacancyRoutes(apiv1 *gin.RouterGroup) {
	candidates := apiv1.Group("/candidates").Use(middlewares.AuthorizationWithBearer())
	{
		candidates.Handle("GET", "/health", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "just to make sure that the 'write-through' service is running",
			})
		})
		candidates.Handle("POST", "/pipelines", handler_candidates.PipelinesWithWriteThrough)
	}

	handler := &handlers.VacancyHandler{}
	admin := apiv1.Use(middlewares.RequestLogs(), middlewares.AuthorizationWithBearer()) // no request logs at first
	{
		admin.Handle("POST", "/vacancies", handler.WriteThroughService)
		admin.Handle("GET", "/vacancies", handler.ReadWriteThroughService)
		admin.Handle("PATCH", "/vacancies", handler.UpdateWriteThroughService)
	}
}
