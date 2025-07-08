package routes

import (
	"go-read-through-service/internal/handlers"
	"go-read-through-service/internal/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

func VacancyRoutes(apiv1 *gin.RouterGroup) {
	public := apiv1.Use(middlewares.PublicIdentityCheck())
	{
		public.Handle("GET", "/public/health", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "just to make sure that the 'read-through' service is running",
			})
		})
		public.Handle("GET", "/public/vacancies", handlers.VacanciesWithReadThrough)
	}

	handler := &handlers.VacancyHandler{}
	admin := apiv1.Use(middlewares.RequestLogs(), middlewares.AuthorizationWithBearer()) // no request logs at first
	{
		// apiv1.Handle("GET", "/vacancies", handler.ReadCacheAsideVacancies)
		admin.Handle("POST", "/vacancies", handler.WriteVacanciesReadThrough)
		admin.Handle("GET", "/vacancies", handler.ReadThroughService)
		admin.Handle("PATCH", "/vacancies", handler.UpdateVacanciesReadThrough)
	}
}
