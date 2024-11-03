package routes

import (
	"future-interns-backend/internal/handlers"
	"future-interns-backend/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func PermissionRoutes(apiv1 *gin.RouterGroup) {
	permissionHandlers := &handlers.PermissionsHandler{}

	router := apiv1.Group("/permissions")
	/* middlewares */
	router.Use(middlewares.AuthorizationWithBearer())
	{
		router.Handle(MethodPost, "/", permissionHandlers.CreatePermissions)
	}
}
