package routes

import (
	"future-interns-backend/internal/handlers"
	"future-interns-backend/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func CandidateRoutes(apiv1 *gin.RouterGroup) {
	candidateHandlers := &handlers.CandidatesHandler{}

	router := apiv1.Group("/candidates")
	router.Use(middlewares.AuthorizationWithBearer())
	{
		router.Handle(MethodPost, "/", candidateHandlers.Create)
		router.Handle(MethodPatch, "/", candidateHandlers.Update)
		router.Handle(MethodDelete, "/:id", candidateHandlers.DeleteById)
		router.Handle(MethodGet, "/", candidateHandlers.Get)
		router.Handle(MethodGet, "/:id", candidateHandlers.GetById)
		router.Handle(MethodGet, "/unscoped", candidateHandlers.Unscoped)
		router.Handle(MethodGet, "/unscoped/:id", candidateHandlers.UnscopedById)
	}
}
