package routes

import (
	"future-interns-backend/internal/constants"
	"future-interns-backend/internal/handlers"
	"future-interns-backend/internal/handlers/tests"
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

	/* developer administrator */
	AdministratorSocials := router.Group("/socials")
	{
		AdministratorSocials.Handle(constants.MethodPost, "/", administratorHandlers.CreateSocial)
	}

	routerVacancies := router.Group("/vacancies")
	{
		routerGenerate := routerVacancies.Group("/generates")
		{
			routerGenerate.Handle(constants.MethodGet, "/", administratorHandlers.GenerateRawVacancies)
			routerGenerate.Handle(constants.MethodPost, "/", administratorHandlers.GenerateVacancies)
			routerGenerate.Handle(constants.MethodDelete, "/:sla", administratorHandlers.DeleteGeneratedVacancies)
		}
	}

	/* Dashboard */
	routerDashboard := router.Group("/dashboard")
	{
		routerDashboard.Handle(constants.MethodGet, "/performances", administratorHandlers.GetPerformanceTestResults)
	}

	routerTest := router.Group("/test")
	{
		routerTest.Handle(constants.MethodPost, "/", administratorHandlers.CreateNewTestSession)
		routerTest.Handle(constants.MethodGet, "/:sessionID/status", administratorHandlers.CheckSessionTestStatus)
		routerTest.Handle(constants.MethodGet, "/:sessionID/logs", administratorHandlers.GetRequestLogsByPattern)
		routerTest.Handle(constants.MethodGet, "/:sessionID/logs/all", administratorHandlers.GetFullRequestLogsBySession)
		routerTest.Handle(constants.MethodGet, "/:sessionID/logs/no-cache", administratorHandlers.GetRequestLogsNoCache)

		routerNoCache := routerTest.Group("/no-cache")
		{
			routerNoCache.Handle(constants.MethodGet, "/:sessionID", administratorHandlers.GetCacheSessionByID)
			routerNoCache.Handle(constants.MethodGet, "/:sessionID/logs", administratorHandlers.GetNoCacheLogs)
		}

		routerGenerate := routerTest.Group("/generates")
		generateHandlers := &tests.GenerateHandler{}
		{
			routerGenerate.Handle(constants.MethodGet, "/sampling", generateHandlers.MakeSampling)
			routerGenerate.Handle(constants.MethodPost, "/vacancies", generateHandlers.MakeFakeVacancies)
			routerGenerate.Handle(constants.MethodDelete, "/vacancies", generateHandlers.DeleteFakeVacancies)
			routerGenerate.Handle(constants.MethodPost, "/vacancies/store", generateHandlers.StoreRawVacancies)
			// routerGenerate.Handle(constants.MethodGet, "/vacancies", administratorHandlers.GenerateRawVacancies)
			routerGenerate.Handle(constants.MethodGet, "/tokens", administratorHandlers.GenerateRandomToken)
		}

	}
}
