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
	routerAddress := router.Group("/addresses")
	{
		routerAddress.Handle(MethodPost, "/", candidateHandlers.StoreAddresses)
		routerAddress.Handle(MethodPatch, "/", candidateHandlers.UpdateAddress)
		routerAddress.Handle(MethodGet, "/:id", candidateHandlers.AddressGetById)
		routerAddress.Handle(MethodDelete, "/:id", candidateHandlers.AddressDeleteById)
	}
	routerEducation := router.Group("/educations")
	{
		routerEducation.Handle(MethodPost, "/", candidateHandlers.StoreEducations)
		routerEducation.Handle(MethodPatch, "/", candidateHandlers.UpdateEducation)
		routerEducation.Handle(MethodGet, "/:id", candidateHandlers.EducationGetById)
		routerEducation.Handle(MethodDelete, "/:id", candidateHandlers.EducationDeleteById)
	}
	routerExperience := router.Group("/experiences")
	{
		routerExperience.Handle(MethodPost, "/", candidateHandlers.StoreExperience)
		routerExperience.Handle(MethodPatch, "/", candidateHandlers.UpdateExperience)
		routerExperience.Handle(MethodGet, "/:id", candidateHandlers.ExperienceGetById)
		routerExperience.Handle(MethodDelete, "/:id", candidateHandlers.ExperienceDeleteById)
	}
	routerSocial := router.Group("/socials")
	{
		routerSocial.Handle(MethodPost, "/", candidateHandlers.StoreCandidateSocial)
		routerSocial.Handle(MethodPatch, "/", candidateHandlers.UpdateCandidateSocial)
		routerSocial.Handle(MethodDelete, "/:socialId", candidateHandlers.CandidateSocialDeleteById)
	}
	routerSkill := router.Group("/skills")
	{
		routerSkill.Handle(MethodPost, "/", candidateHandlers.StoreCandidateSkill)
		routerSkill.Handle(MethodDelete, "/:skillId", candidateHandlers.CandidateSkillDeleteById)
	}
}
