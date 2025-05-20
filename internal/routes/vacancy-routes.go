package routes

import (
	"go-no-cache-service/internal/handlers"
	"go-no-cache-service/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func VacancyRoutes(apiv1 *gin.RouterGroup) {
	handler := &handlers.VacancyHandler{}
	apiv1.Use(middlewares.RequestLogs(), middlewares.AuthorizationWithBearer()) // no request logs at first
	{
		apiv1.Handle("GET", "/vacancies", handler.GetVacanciesNoCache)
		apiv1.Handle("POST", "/vacancies", handler.WriteVacanciesNoCache)
		// apiv1.Handle("GET", "/vacancies/write-ops", handler.ReadVacanciesNoCache)
		apiv1.Handle("PATCH", "/vacancies", handler.UpdateVacanciesNoCache)
	}
}
