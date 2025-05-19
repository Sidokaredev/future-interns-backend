package routes

import (
	"go-write-behind-service/internal/handlers"
	"go-write-behind-service/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func VacancyRoutes(apiv1 *gin.RouterGroup) {
	handler := &handlers.VacancyHandler{}
	apiv1.Use(middlewares.RequestLogs(), middlewares.AuthorizationWithBearer()) // no request logs at first
	{
		apiv1.Handle("POST", "/vacancies", handler.WriteBehindService)
		apiv1.Handle("PATCH", "/vacancies", handler.UpdateWriteBehindService)
		apiv1.Handle("GET", "/vacancies", handler.ReadWriteBehindService)
		apiv1.Handle("GET", "/job-status", handler.BackgroundJobStatus)
	}
}
