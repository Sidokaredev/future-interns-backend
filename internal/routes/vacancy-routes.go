package routes

import (
	"go-write-through-service/internal/handlers"
	"go-write-through-service/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func VacancyRoutes(apiv1 *gin.RouterGroup) {
	handler := &handlers.VacancyHandler{}
	apiv1.Use(middlewares.RequestLogs(), middlewares.AuthorizationWithBearer()) // no request logs at first
	{
		apiv1.Handle("POST", "/vacancies", handler.WriteThroughService)
		apiv1.Handle("GET", "/vacancies", handler.ReadWriteThroughService)
		apiv1.Handle("PATCH", "/vacancies", handler.UpdateWriteThroughService)
	}
}
