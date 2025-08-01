package middlewares

import (
	"fmt"
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/handlers"
	"future-interns-backend/internal/models"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RoleCheck() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		bearerToken := strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer ")
		claims := handlers.ParseJWT(bearerToken)

		gormDB, _ := initializer.GetGorm()
		identityAccess := map[string]interface{}{}
		permissions := []map[string]interface{}{}

		GetUserRole := gormDB.Model(&models.IdentityAccess{}).
			Select([]string{
				"roles.id as role_id",
				"roles.name",
				// "roles.description",
				"identity_accesses.type",
			}).Joins("INNER JOIN roles ON roles.id = identity_accesses.role_id").
			Where("user_id = ?", claims.Id).First(&identityAccess) // DEFAULT IS Find() return multiple identity_accesses

		if GetUserRole.RowsAffected == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   fmt.Errorf("%v identity found", GetUserRole.RowsAffected),
				"message": "unidentified user, your account has no access to any resources",
			})

			ctx.Abort()
			return
		}

		// listOfIdentity := []string{}
		// for _, identity := range identityAccess {
		// 	listOfIdentity = append(listOfIdentity, identity["name"].(string))
		// }
		// useIdentity := map[string]interface{}{}

		// rolesId := []uint{}
		// for _, identity := range identityAccess {
		// 	rolesId = append(rolesId, identity["id"].(uint))
		// }

		roleID := identityAccess["role_id"]
		delete(identityAccess, "role_id")

		GetPermissions := gormDB.Model(&models.RolePermission{}).
			Select([]string{
				"permissions.name",
				// "permissions.resource",
				// "permissions.description",
			}).
			Joins("INNER JOIN permissions ON permissions.id = role_permissions.permission_id").
			Where("role_id = ?", roleID).
			Find(&permissions)

		if GetPermissions.RowsAffected == 0 {
			log.Println("no permissions exists for this identity")
		}

		listOfPermissions := map[string]bool{}
		for _, permission := range permissions {
			listOfPermissions[permission["name"].(string)] = true
		}

		ctx.Set("identity-accesses", identityAccess)
		ctx.Set("permissions", listOfPermissions)
		ctx.Set("user-id", claims.Id) // NEW -> string -> userID
		ctx.Next()
	}
}
