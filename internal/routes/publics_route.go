package routes

import (
	"future-interns-backend/internal/constants"
	"future-interns-backend/internal/handlers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PublicRoutes(apiv1 *gin.RouterGroup) {
	publicHandlers := &handlers.PublicHandlers{}

	routePublic := apiv1.Group("/public")

	routeSkill := routePublic.Group("/skills")
	{
		routeSkill.Handle(http.MethodGet, "/", publicHandlers.GetSkills)
	}
	routeSocial := routePublic.Group("/socials")
	{
		routeSocial.Handle(constants.MethodGet, "/", publicHandlers.GetSocials)
	}
}
