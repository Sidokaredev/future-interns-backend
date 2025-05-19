package routes

import "github.com/gin-gonic/gin"

func LoadRoutes(engine *gin.Engine) {
	/* Global Middleware */

	/* versioning api v1 */
	apiv1 := engine.Group("/api/v1")
	AccountsRoutes(apiv1)
	CandidateRoutes(apiv1)
	EmployerRoutes(apiv1)
	AdministratorRoutes(apiv1)
	VacancyRoutes(apiv1)
	ImageRoutes(apiv1)
	RoleRoutes(apiv1)
	PermissionRoutes(apiv1)
	PublicRoutes(apiv1)
	ContainersRoute(apiv1)
}
