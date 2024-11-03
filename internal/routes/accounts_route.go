package routes

import (
	"future-interns-backend/internal/handlers"
	"future-interns-backend/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func AccountsRoutes(apiv1 *gin.RouterGroup) {
	accountHandlers := &handlers.AccountsHandler{}

	router := apiv1.Group("/accounts")
	/* use middleware here */
	{
		router.Handle(MethodPost, "/auth", accountHandlers.Auth)
		router.Handle(MethodPost, "/create", accountHandlers.RegisterAccount)
		router.Handle(MethodGet, "/identities", func(ctx *gin.Context) {})
		router.Handle(MethodGet, "/identities/:id", func(ctx *gin.Context) {})
		router.Handle(MethodGet, "/getUser/:userId", middlewares.AuthorizationWithBearer(), accountHandlers.GetUser)
	}
}
