package routes

import (
	"future-interns-backend/internal/constants"
	"future-interns-backend/internal/handlers"
	"future-interns-backend/internal/middlewares"

	"github.com/gin-gonic/gin"
)

func EmployerRoutes(apiv1 *gin.RouterGroup) {
	employerHandlers := &handlers.EmployerHandlers{}

	router := apiv1.Group("/employers")
	// middlewares
	router.Use(middlewares.AuthorizationWithBearer(), middlewares.RoleCheck())
	{
		router.Handle(constants.MethodPost, "/", employerHandlers.StoreEmployer)
		router.Handle(constants.MethodPatch, "/", employerHandlers.UpdateEmployer)
		router.Handle(constants.MethodGet, "/", employerHandlers.GetEmployer)
		router.Handle(constants.MethodDelete, "/", employerHandlers.DeleteEmployer)
		router.Handle(constants.MethodGet, "/check", employerHandlers.CheckProfile)
	}
	routerHeadquarter := router.Group("/headquarters")
	{
		routerHeadquarter.Handle(constants.MethodPost, "/", employerHandlers.StoreHeadquarter)
		routerHeadquarter.Handle(constants.MethodPatch, "/:id", employerHandlers.UpdateHeadquarter)
		routerHeadquarter.Handle(constants.MethodGet, "/", employerHandlers.ListHeadquarter)
		routerHeadquarter.Handle(constants.MethodGet, "/:id", employerHandlers.GetHeadquarter)
		routerHeadquarter.Handle(constants.MethodDelete, "/:id", employerHandlers.DeleteHeadquarter)
	}
	routerOfficeImages := router.Group("/office-images")
	{
		routerOfficeImages.Handle(constants.MethodPost, "/", employerHandlers.StoreOfficeImage)
		routerOfficeImages.Handle(constants.MethodGet, "/", employerHandlers.GetOfficeImages)
		routerOfficeImages.Handle(constants.MethodPatch, "/:id", employerHandlers.UpdateOfficeImage)
		routerOfficeImages.Handle(constants.MethodDelete, "/:id", employerHandlers.DeleteOfficeImage)
	}
	routerSocial := router.Group("/socials")
	{
		routerSocial.Handle(constants.MethodPost, "/", employerHandlers.StoreEmployerSocials)
		routerSocial.Handle(constants.MethodGet, "/", employerHandlers.GetEmployerSocials)
		routerSocial.Handle(constants.MethodPatch, "/", employerHandlers.UpdateEmployerSocial)
		routerSocial.Handle(constants.MethodDelete, "/:id", employerHandlers.DeleteEmployerSocial)
	}
	routerVacancies := router.Group("/vacancies")
	{
		routerVacancies.Handle(constants.MethodPost, "/", employerHandlers.StoreVacancy)
		routerVacancies.Handle(constants.MethodPatch, "/:id", employerHandlers.UpdateVacancy)
		routerVacancies.Handle(constants.MethodGet, "/:id", employerHandlers.GetEmployer)
		routerVacancies.Handle(constants.MethodDelete, "/:id", employerHandlers.DeleteVacancy)
		routerVacancies.Handle(constants.MethodGet, "/", employerHandlers.ListVacancy)
	}
	// routerScreening := router.Group("/screenings")
	{
		// routerScreening.Handle(constants.MethodGet, "/", employerHandlers.UpdateS)
	}
	routerAssessment := router.Group("/assessments")
	{
		routerAssessment.Handle(constants.MethodPost, "/", employerHandlers.StoreAssessment)
		routerAssessment.Handle(constants.MethodPatch, "/:id", employerHandlers.UpdateAssessment)
		routerAssessment.Handle(constants.MethodGet, "/:id", employerHandlers.ListAssessment)
		// routerAssessment.Handle(contants.MethodGet, "/:id", employerHandlers.GetAssessment)
		routerAssessment.Handle(constants.MethodDelete, "/:assessmentID", employerHandlers.DeleteAssessment)
		routerAssessment.Handle(constants.MethodDelete, "/:assessmentID/assessment-document/:documentID", employerHandlers.DeleteAssessmentDocument)

		routerAssessmentAssignee := routerAssessment.Group("/assignees")
		{
			routerAssessmentAssignee.Handle(constants.MethodPost, "/", employerHandlers.StoreAssignees)
			routerAssessmentAssignee.Handle(constants.MethodPatch, "/", employerHandlers.UpdateAssignee)
			routerAssessmentAssignee.Handle(constants.MethodDelete, "/:assessmentId/:pipelineId", employerHandlers.DeleteAssignee)
		}
	}
	routerInterview := router.Group("/interviews")
	{
		routerInterview.Handle(constants.MethodPost, "/", employerHandlers.StoreInterview)
		routerInterview.Handle(constants.MethodPatch, "/:id", employerHandlers.UpdateInterview)
		routerInterview.Handle(constants.MethodGet, "/histories/:pipelineId/:vacancyId", employerHandlers.ListInterviewHistory)
		routerInterview.Handle(constants.MethodDelete, "/:id", employerHandlers.DeleteInterview)
	}
	routerOffering := router.Group("/offerings")
	{
		routerOffering.Handle(constants.MethodPost, "/", employerHandlers.StoreOffering)
		routerOffering.Handle(constants.MethodPatch, "/:id", employerHandlers.UpdateOffering)
		// routerOffering.Handle(constants.MethodGet, "/:id", employerHandlers.GetOffering)
		routerOffering.Handle(constants.MethodDelete, "/:id", employerHandlers.DeleteOffering)
		routerOffering.Handle(constants.MethodGet, "/:id", employerHandlers.ListOffering)
	}
	routerPipeline := router.Group("/pipelines")
	{
		routerPipeline.Handle(constants.MethodPatch, "/", employerHandlers.UpdatePipeline)
		routerPipeline.Handle(constants.MethodGet, "/:vacancyID/screening", employerHandlers.GetApplicantScreening)
		routerPipeline.Handle(constants.MethodGet, "/:vacancyID/assessment", employerHandlers.GetApplicantAssessment)
		routerPipeline.Handle(constants.MethodGet, "/:vacancyID/interview", employerHandlers.GetApplicantInterview)
		routerPipeline.Handle(constants.MethodGet, "/:vacancyID/offering", employerHandlers.GetApplicantsOffering)
	}
}
