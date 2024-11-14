package routes

import (
	"future-interns-backend/internal/handlers"
	"future-interns-backend/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func EmployerRoutes(apiv1 *gin.RouterGroup) {
	employerHandlers := &handlers.EmployerHandlers{}

	router := apiv1.Group("/employers")
	// use middlewares here
	router.Use(middlewares.AuthorizationWithBearer(), middlewares.RoleCheck())
	{
		router.Handle(MethodPost, "/", employerHandlers.StoreEmployer)
		router.Handle(MethodPatch, "/", employerHandlers.UpdateEmployer)
		router.Handle(MethodGet, "/", employerHandlers.GetEmployer)
		router.Handle(MethodDelete, "/", employerHandlers.DeleteEmployer)
	}
	routerHeadquarter := router.Group("/headquarters")
	// middlewares here
	{
		routerHeadquarter.Handle(MethodPost, "/", employerHandlers.StoreHeadquarter)
		routerHeadquarter.Handle(MethodPatch, "/:id", employerHandlers.UpdateHeadquarter)
		routerHeadquarter.Handle(MethodGet, "/", employerHandlers.ListHeadquarter)
		routerHeadquarter.Handle(MethodGet, "/:id", employerHandlers.GetHeadquarter)
		routerHeadquarter.Handle(MethodDelete, "/:id", employerHandlers.DeleteHeadquarter)
	}
	routerOfficeImages := router.Group("/office-images")
	{
		routerOfficeImages.Handle(MethodPost, "/", employerHandlers.StoreOfficeImage)
		routerOfficeImages.Handle(MethodPatch, "/:id", employerHandlers.UpdateOfficeImage)
		routerOfficeImages.Handle(MethodDelete, "/:id", employerHandlers.DeleteOfficeImage)
	}
	routerSocial := router.Group("/socials")
	{
		routerSocial.Handle(MethodPost, "/", employerHandlers.StoreEmployerSocials)
		routerSocial.Handle(MethodPatch, "/:id", employerHandlers.UpdateEmployerSocial)
		routerSocial.Handle(MethodDelete, "/:id", employerHandlers.DeleteEmployerSocial)
	}
}
