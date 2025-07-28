package routes

import (
	"go-cache-aside-service/internal/handlers"
	"go-cache-aside-service/internal/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

func VacancyRoutes(apiv1 *gin.RouterGroup) {
	public := apiv1.Use(middlewares.PublicIdentityCheck())
	{
		public.Handle("GET", "/public/health", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "just to make sure that the 'cache-aside' service is running",
			})
		})
		public.Handle("GET", "/public/vacancies", handlers.VacanciesWithCacheAside)
		public.Handle("GET", "/public/vacancies/:id", handlers.VacanciesByIdWithCacheAside)
	}

	handler := &handlers.VacancyHandler{}
	admin := apiv1.Use(middlewares.RequestLogs(), middlewares.AuthorizationWithBearer())
	{
		admin.Handle("POST", "/vacancies", handler.WriteVacanciesCacheAside)
		admin.Handle("GET", "/vacancies", handler.ReadCacheAsideService)
		admin.Handle("PATCH", "/vacancies", handler.UpdateVacanciesCacheAside)
	}
}
