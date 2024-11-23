package routes

import (
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
		router.Handle(MethodPost, "/", employerHandlers.StoreEmployer)
		router.Handle(MethodPatch, "/", employerHandlers.UpdateEmployer)
		router.Handle(MethodGet, "/", employerHandlers.GetEmployer)
		router.Handle(MethodDelete, "/", employerHandlers.DeleteEmployer)
	}
	routerHeadquarter := router.Group("/headquarters")
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
	routerVacancies := router.Group("/vacancies")
	{
		routerVacancies.Handle(MethodPost, "/", employerHandlers.StoreVacancy)
		routerVacancies.Handle(MethodPatch, "/:id", employerHandlers.UpdateVacancy)
		routerVacancies.Handle(MethodGet, "/:id", employerHandlers.GetEmployer)
		routerVacancies.Handle(MethodDelete, "/:id", employerHandlers.DeleteVacancy)
		routerVacancies.Handle(MethodGet, "/", employerHandlers.ListVacancy)
	}
	routerAssessment := router.Group("/assessments")
	{
		routerAssessment.Handle(MethodPost, "/", employerHandlers.StoreAssessment)
		routerAssessment.Handle(MethodPatch, "/:id", employerHandlers.UpdateAssessment)
		routerAssessment.Handle(MethodGet, "/:id", employerHandlers.ListAssessment)
		// routerAssessment.Handle(MethodGet, "/:id", employerHandlers.GetAssessment)
		routerAssessment.Handle(MethodDelete, "/:id", employerHandlers.DeleteAssessment)

		routerAssessmentAssignee := routerAssessment.Group("/assignees")
		{
			routerAssessmentAssignee.Handle(MethodPost, "/", employerHandlers.StoreAssignees)
			routerAssessmentAssignee.Handle(MethodPatch, "/", employerHandlers.UpdateAssignee)
			routerAssessmentAssignee.Handle(MethodDelete, "/:assessmentId/:pipelineId", employerHandlers.DeleteAssignee)
		}
	}
	routerInterview := router.Group("/interviews")
	{
		routerInterview.Handle(MethodPost, "/", employerHandlers.StoreInterview)
		routerInterview.Handle(MethodPatch, "/:id", employerHandlers.UpdateInterview)
		routerInterview.Handle(MethodGet, "/histories/:pipelineId/:vacancyId", employerHandlers.ListInterviewHistory)
		routerInterview.Handle(MethodDelete, "/:id", employerHandlers.DeleteInterview)
	}
	routerOffering := router.Group("/offerings")
	{
		routerOffering.Handle(MethodPost, "/", employerHandlers.StoreOffering)
		routerOffering.Handle(MethodPatch, "/:id", employerHandlers.UpdateOffering)
		// routerOffering.Handle(MethodGet, "/:id", employerHandlers.GetOffering)
		routerOffering.Handle(MethodDelete, "/:id", employerHandlers.DeleteOffering)
		routerOffering.Handle(MethodGet, "/:id", employerHandlers.ListOffering)
	}
}
