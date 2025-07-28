package middlewares

import (
	"fmt"
	initializer "go-cache-aside-service/init"
	"go-cache-aside-service/internal/models"
	"log"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

func PublicIdentityCheck() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var authenticated bool
		var identity map[string]any
		var permissions []string
		var userID string

		bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
		claims := TokenClaims{}
		if bearerToken != "" { // -> cek apakah request menyertakan Bearer Token
			var secretKey = []byte(viper.GetString("authorization.jwt.secretKey"))
			_, errParse := jwt.ParseWithClaims(bearerToken, &claims, func(token *jwt.Token) (any, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return secretKey, nil
			})

			if errParse == nil { // -> jika token berhasil diverifikasi (tidak terdapat error)
				authenticated = true // -> mengubah variabel [authenticated] 'true'
				userID = claims.Id   // -> paylod token {'Id': 'user ID'}

				gormDB, _ := initializer.GetMssqlDB()
				errGetIdentity := gormDB.Model(&models.IdentityAccess{}).Select([]string{
					"identity_accesses.role_id",
					"identity_accesses.type",
					"roles.name",
				}).
					Joins("INNER JOIN roles ON roles.id = identity_accesses.role_id").
					Where("user_id = ?", claims.Id).First(&identity).Error

				if errGetIdentity == nil { // -> jika pengguna memiliki identitas akses (tidak terdapat error)
					getPermissions := gormDB.Model(&models.RolePermission{}).
						Select("permissions.name").
						Joins("INNER JOIN permissions ON permissions.id = role_permissions.permission_id").
						Where("role_id = ?", identity["role_id"]).Find(&permissions)

					if getPermissions.RowsAffected == 0 {
						log.Printf("Middleware:public-identity-check say -> tidak terdapat hak akses untuk [%v]", identity["type"])
					}
				}
			}
		}

		ctx.Set("authenticated", authenticated) // bool
		ctx.Set("identity", identity["type"])   // map[string]any - keys -> [name, type]
		ctx.Set("permissions", permissions)     // []string
		ctx.Set("token", bearerToken)           // string
		ctx.Set("user-id", userID)              // string

		ctx.Next()
	}
}
