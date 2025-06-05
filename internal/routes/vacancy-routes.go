package routes

import (
	"go-cache-aside-service/internal/handlers"
	"go-cache-aside-service/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func VacancyRoutes(apiv1 *gin.RouterGroup) {
	handler := &handlers.VacancyHandler{}
	apiv1.Use(middlewares.RequestLogs(), middlewares.AuthorizationWithBearer())
	{
		apiv1.Handle("POST", "/vacancies", handler.WriteVacanciesCacheAside)
		apiv1.Handle("GET", "/vacancies", handler.ReadCacheAsideService)
		apiv1.Handle("PATCH", "/vacancies", handler.UpdateVacanciesCacheAside)
	}
}
