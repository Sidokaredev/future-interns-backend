package routes

import (
	"future-interns-backend/internal/constants"
	"future-interns-backend/internal/handlers"
	"future-interns-backend/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func CandidateRoutes(apiv1 *gin.RouterGroup) {
	candidateHandlers := &handlers.CandidatesHandler{}

	router := apiv1.Group("/candidates")
	router.Use(middlewares.AuthorizationWithBearer(), middlewares.RoleCheck())
	{
		router.Handle(constants.MethodPost, "/", candidateHandlers.Create)
		router.Handle(constants.MethodPatch, "/", candidateHandlers.Update)
		router.Handle(constants.MethodDelete, "/:id", candidateHandlers.DeleteById)
		router.Handle(constants.MethodGet, "/", candidateHandlers.Get)
		router.Handle(constants.MethodGet, "/:id", candidateHandlers.GetById)
		router.Handle(constants.MethodGet, "/unscoped", candidateHandlers.Unscoped)
		router.Handle(constants.MethodGet, "/unscoped/:id", candidateHandlers.UnscopedById)
		router.Handle(constants.MethodGet, "/user", candidateHandlers.UserGet)
		router.Handle(constants.MethodGet, "/check", candidateHandlers.CheckProfile)
	}
	routerAddress := router.Group("/addresses")
	{
		routerAddress.Handle(constants.MethodPost, "/", candidateHandlers.StoreAddresses)
		routerAddress.Handle(constants.MethodPatch, "/", candidateHandlers.UpdateAddress)
		routerAddress.Handle(constants.MethodGet, "/", candidateHandlers.AddressGet)
		routerAddress.Handle(constants.MethodGet, "/:id", candidateHandlers.AddressGetById)
		routerAddress.Handle(constants.MethodDelete, "/:id", candidateHandlers.AddressDeleteById)
	}
	routerEducation := router.Group("/educations")
	{
		routerEducation.Handle(constants.MethodPost, "/", candidateHandlers.StoreEducations)
		routerEducation.Handle(constants.MethodPatch, "/", candidateHandlers.UpdateEducation)
		routerEducation.Handle(constants.MethodGet, "/", candidateHandlers.EducationsGet)
		routerEducation.Handle(constants.MethodGet, "/:id", candidateHandlers.EducationGetById)
		routerEducation.Handle(constants.MethodDelete, "/:id", candidateHandlers.EducationDeleteById)
	}
	routerExperience := router.Group("/experiences")
	{
		routerExperience.Handle(constants.MethodPost, "/", candidateHandlers.StoreExperience)
		routerExperience.Handle(constants.MethodPatch, "/", candidateHandlers.UpdateExperience)
		routerExperience.Handle(constants.MethodGet, "/", candidateHandlers.ExperiencesGet)
		routerExperience.Handle(constants.MethodGet, "/:id", candidateHandlers.ExperienceGetById)
		routerExperience.Handle(constants.MethodDelete, "/:id", candidateHandlers.ExperienceDeleteById)
	}
	routerSocial := router.Group("/socials")
	{
		routerSocial.Handle(constants.MethodPost, "/", candidateHandlers.StoreCandidateSocial)
		routerSocial.Handle(constants.MethodPatch, "/", candidateHandlers.UpdateCandidateSocial)
		routerSocial.Handle(constants.MethodGet, "/", candidateHandlers.SocialsGet)
		routerSocial.Handle(constants.MethodDelete, "/:socialId", candidateHandlers.CandidateSocialDeleteById)
	}
	routerSkill := router.Group("/skills")
	{
		routerSkill.Handle(constants.MethodPost, "/", candidateHandlers.StoreCandidateSkill)
		routerSkill.Handle(constants.MethodGet, "/", candidateHandlers.SkillsGet)
		routerSkill.Handle(constants.MethodDelete, "/:skillId", candidateHandlers.CandidateSkillDeleteById)
	}
	routerPipeline := router.Group("/pipelines")
	{
		routerPipeline.Handle(constants.MethodPost, "/", candidateHandlers.CreatePipeline)
		routerPipeline.Handle(constants.MethodGet, "/", candidateHandlers.ListPipeline)
		routerPipeline.Handle(constants.MethodGet, "/:pipelineID/assessments/:vacancyID", candidateHandlers.GetPipelineAssessments)
		routerPipeline.Handle(constants.MethodGet, "/:pipelineID/vacancies/:vacancyID/interviews", candidateHandlers.GetPipelineInterviews)
		routerPipeline.Handle(constants.MethodGet, "/:pipelineID/vacancies/:vacancyID/offering", candidateHandlers.GetPipelineOfferings)
	}
	routerAssessment := router.Group("/assessments")
	{
		// all stuff about assessments here
	}
	routerAssessmentSubmission := routerAssessment.Group("/submissions")
	{
		routerAssessmentSubmission.Handle(constants.MethodPost, "/", candidateHandlers.StoreAssessmentSubmissions)
		routerAssessmentSubmission.Handle(constants.MethodDelete, "/:id", candidateHandlers.DeleteAssessmentSubmission)
	}
	// routerInterview := router.Group("/interviews")
	// {
	// 	routerInterview.Handle(constants.MethodGet, "/", candidateHandlers.GetPipelineInterviews)
	// }
	routerOffering := router.Group("/offerings")
	{
		routerOffering.Handle(constants.MethodPatch, "/:id", candidateHandlers.UpdateOffering)
	}
}
