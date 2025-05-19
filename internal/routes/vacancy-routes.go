package routes

import (
	"go-read-through-service/internal/handlers"
	"go-read-through-service/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func VacancyRoutes(apiv1 *gin.RouterGroup) {
	handler := &handlers.VacancyHandler{}
	apiv1.Use(middlewares.RequestLogs(), middlewares.AuthorizationWithBearer()) // no request logs at first
	{
		// apiv1.Handle("GET", "/vacancies", handler.ReadCacheAsideVacancies)
		apiv1.Handle("POST", "/vacancies", handler.WriteVacanciesReadThrough)
		apiv1.Handle("GET", "/vacancies", handler.ReadThroughService)
		apiv1.Handle("PATCH", "/vacancies", handler.UpdateVacanciesReadThrough)
	}
}
