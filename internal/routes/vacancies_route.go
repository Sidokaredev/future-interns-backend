package routes

import (
	"future-interns-backend/internal/constants"
	"future-interns-backend/internal/handlers"
	"future-interns-backend/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func VacancyRoutes(apiv1 *gin.RouterGroup) {
	vacancyHandlers := handlers.VacancyHandlers{}

	router := apiv1.Group("/vacancies")
	/* middlewares */
	router.Use(middlewares.AuthorizationWithBearer())
	{
		router.Handle(constants.MethodGet, "/", vacancyHandlers.GetVacancies)
	}
}
