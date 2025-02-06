package routes

import (
	"go-read-through-service/internal/handlers"
	"go-read-through-service/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func VacancyRoutes(apiv1 *gin.RouterGroup) {
	handler := &handlers.VacancyHandler{}
	apiv1.Use(middlewares.AuthorizationWithBearer()) // no request logs at first
	{
		apiv1.Handle("GET", "/vacancies", handler.GetVacanciesReadThrough)
	}
}
