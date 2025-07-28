package middlewares

import (
	"fmt"
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/handlers"
	"future-interns-backend/internal/models"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

func PublicIdentityCheck() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var authenticated bool
		var identity map[string]interface{}
		var permissions []string
		var userID string

		bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
		claims := handlers.TokenClaims{}
		if bearerToken != "" { // if bearer token exist
			var secretKey = []byte(viper.GetString("authorization.jwt.secretKey"))
			_, errParse := jwt.ParseWithClaims(bearerToken, &claims, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return secretKey, nil
			})

			// log.Println("error parse \t:", errParse.Error())

			if errParse == nil { // if parse token go for successful
				authenticated = true // set authenticated "true" after pass token parse
				userID = claims.Id

				gormDB, _ := initializer.GetGorm()
				errGetIdentity := gormDB.Model(&models.IdentityAccess{}).Select([]string{
					"role_id",
					"type",
				}).Where("user_id = ?", claims.Id).First(&identity).Error

				if errGetIdentity == nil { // if getting identity faced no error
					getPermissions := gormDB.Model(&models.RolePermission{}).
						Select("permissions.name").
						Joins("INNER JOIN permissions ON permissions.id = role_permissions.permission_id").
						Where("role_id = ?", identity["role_id"]).Find(&permissions)

					if getPermissions.RowsAffected == 0 {
						log.Println("public_identity_check Middleware says \t: user has no permissions")
					}
				}
			}
		}

		ctx.Set("authenticated", authenticated)      // bool
		ctx.Set("identity-access", identity["type"]) // string
		ctx.Set("permissions", permissions)          // slice
		ctx.Set("token", bearerToken)                // string
		ctx.Set("user-id", userID)                   // string

		ctx.Next()
	}
}
