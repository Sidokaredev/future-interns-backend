package routes

import (
	"future-interns-backend/internal/handlers"

	"github.com/gin-gonic/gin"
)

func AccountsRoutes(apiv1 *gin.RouterGroup) {
	accountHandlers := &handlers.AccountsHandler{}

	router := apiv1.Group("/accounts")
	/* use middleware here */
	{
		router.Handle(MethodPost, "/auth", accountHandlers.Auth)
		router.Handle(MethodPost, "/create", accountHandlers.RegisterAccount)
	}
}
