package routes

import "github.com/gin-gonic/gin"

const (
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodDelete  = "DELETE"
	MethodPatch   = "PATCH"
	MethodHead    = "HEAD"
	MethodOptions = "OPTIONS"
)

func LoadRoutes(engine *gin.Engine) {
	/* Global Middleware */

	/* versioning api v1 */
	apiv1 := engine.Group("/api/v1")
	AccountsRoutes(apiv1)
	VacancyRoutes(apiv1)
}
