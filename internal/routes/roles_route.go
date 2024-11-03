package routes

import (
	"future-interns-backend/internal/handlers"
	"future-interns-backend/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func RoleRoutes(apiv1 *gin.RouterGroup) {
	rolesHandler := &handlers.RolesHandler{}

	router := apiv1.Group("/roles")
	/* middlewares */
	router.Use(middlewares.AuthorizationWithBearer())
	{
		router.Handle(MethodPost, "/", rolesHandler.CreateRoles)
	}
}
