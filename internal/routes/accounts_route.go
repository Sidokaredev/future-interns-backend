package routes

import (
	"future-interns-backend/internal/constants"
	"future-interns-backend/internal/handlers"
	"future-interns-backend/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func AccountsRoutes(apiv1 *gin.RouterGroup) {
	accountHandlers := &handlers.AccountsHandler{}

	router := apiv1.Group("/accounts")
	/* use middleware here */
	{
		router.Handle(constants.MethodPost, "/auth", accountHandlers.Auth)
		router.Handle(constants.MethodPost, "/create", accountHandlers.RegisterAccount)
		router.Handle(constants.MethodGet, "/identities", func(ctx *gin.Context) {})
		router.Handle(constants.MethodGet, "/identities/:id", func(ctx *gin.Context) {})
		router.Handle(constants.MethodGet, "/user-information", middlewares.AuthorizationWithBearer(), middlewares.RoleCheck(), accountHandlers.UserInformation)
	}
}
