package routes

import (
	"future-interns-backend/internal/constants"
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
		EmployerUser.Handle(constants.MethodPost, "/", administratorHandlers.CreateEmployerUser)
		EmployerUser.Handle(constants.MethodPatch, "/", administratorHandlers.UpdateEmployerUser)
		EmployerUser.Handle(constants.MethodGet, "/:id", administratorHandlers.GetEmployerUserById)
		EmployerUser.Handle(constants.MethodDelete, "/:id", administratorHandlers.DeleteEmployerUserById)
		EmployerUser.Handle(constants.MethodGet, "/", administratorHandlers.ListEmployerUsers)
	}
	EmployerSkill := router.Group("/skills")
	{
		EmployerSkill.Handle(constants.MethodPost, "/", administratorHandlers.CreateSkills)
		EmployerSkill.Handle(constants.MethodDelete, "/:id", administratorHandlers.DeleteSkills)
	}
}
