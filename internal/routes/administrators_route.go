package routes

import (
	"future-interns-backend/internal/handlers"
	"future-interns-backend/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func AdministratorRoutes(apiv1 *gin.RouterGroup) {
	administratorHandlers := &handlers.AdministratorHandlers{}

	router := apiv1.Group("/administrators")
	router.Use(middlewares.AuthorizationWithBearer(), middlewares.RoleCheck())
	/* users */
	routerUser := router.Group("/users")
	/* users -> employers */
	// use middlewares here
	EmployerUser := routerUser.Group("/employers")
	{
		EmployerUser.Handle(MethodPost, "/", administratorHandlers.CreateEmployerUser)
		EmployerUser.Handle(MethodPatch, "/", administratorHandlers.UpdateEmployerUser)
		EmployerUser.Handle(MethodGet, "/:id", administratorHandlers.GetEmployerUserById)
		EmployerUser.Handle(MethodDelete, "/:id", administratorHandlers.DeleteEmployerUserById)
		EmployerUser.Handle(MethodGet, "/", administratorHandlers.ListEmployerUsers)
	}
}
