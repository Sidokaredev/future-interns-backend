package routes

import (
	"go-write-behind-service/internal/handlers"
	handler_candidates "go-write-behind-service/internal/handlers/candidates"
	"go-write-behind-service/internal/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

func VacancyRoutes(apiv1 *gin.RouterGroup) {
	candidates := apiv1.Group("/candidates").Use(middlewares.AuthorizationWithBearer())
	{
		candidates.Handle("GET", "/health", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "just to make sure that the 'write-behind' service is running",
			})
		})
		candidates.Handle("POST", "/pipelines", handler_candidates.PipelinesWithWriteBehind)
	}

	handler := &handlers.VacancyHandler{}
	admin := apiv1.Use(middlewares.RequestLogs(), middlewares.AuthorizationWithBearer()) // no request logs at first
	{
		admin.Handle("POST", "/vacancies", handler.WriteBehindService)
		admin.Handle("PATCH", "/vacancies", handler.UpdateWriteBehindService)
		admin.Handle("GET", "/vacancies", handler.ReadWriteBehindService)
		admin.Handle("GET", "/job-status", handler.BackgroundJobStatus)
	}
}
