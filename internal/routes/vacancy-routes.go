package routes

import (
	"go-cache-aside-service/internal/handlers"
	"go-cache-aside-service/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func VacancyRoutes(apiv1 *gin.RouterGroup) {
	handler := &handlers.VacancyHandler{}
	apiv1.Use(middlewares.AuthorizationWithBearer())
	{
		apiv1.Handle("GET", "/vacancies", handler.GetVacanciesCacheAside)
	}
}
