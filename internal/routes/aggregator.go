package routes

import "github.com/gin-gonic/gin"

func ListenToRoutes(engine *gin.Engine) {
	router := engine.Group("/api/v1/no-cache")
	{
		VacancyRoutes(router)
	}
}
