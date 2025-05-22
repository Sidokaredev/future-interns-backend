package middlewares

import (
	"future-interns-backend/internal/constants"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSPolicy() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		whitelisted := []string{
			"http://localhost:5173",
			"https://sidokaredev.github.io",
			"https://sidokaredev.space",
			"https://e-career.polije.sidokaredev.space",
			"http://192.168.144.152",
		}
		for _, host := range whitelisted {
			if origin == host {
				ctx.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
		}
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")

		if ctx.Request.Method == constants.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()
	}
}
