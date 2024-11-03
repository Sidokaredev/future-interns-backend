package handlers

import (
	"fmt"
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type RolesHandler struct {
}

func (r *RolesHandler) CreateRoles(context *gin.Context) {
	_, exists := context.GetQuery("withGrantPermissions") // check Query Parameter
	switch {
	case exists:
		GrantPermissions(context) // do grant permissions to role defined in request
		return
	default:
		CreateMultipleRoles(context) // just creating roles
		return
	}

}

func CreateMultipleRoles(context *gin.Context) {
	var roles struct {
		Roles []models.Role `json:"roles" binding:"required"`
	}

	if errBind := context.ShouldBindJSON(&roles); errBind != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your JSON fields, kids",
		})
		context.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	createRoles := gormDB.Create(&roles.Roles)
	if createRoles.Error != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   createRoles.Error.Error(),
			"message": "failed to store multiple roles",
		})
		context.Abort()
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    strconv.Itoa(int(createRoles.RowsAffected)) + " roles inserted successfully",
	})
}

func GrantPermissions(context *gin.Context) {
	var rolesGrantPermissions struct {
		RoleId      uint   `json:"role_id" binding:"required"`
		Permissions []uint `json:"permissions" binding:"required"`
	}

	if errBind := context.ShouldBindJSON(&rolesGrantPermissions); errBind != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your JSON fields, kids",
		})
		context.Abort()
		return
	}
	var roleWithPermissions []models.RolePermission
	for _, permissionId := range rolesGrantPermissions.Permissions {
		roleWithPermissions = append(roleWithPermissions, models.RolePermission{
			RoleId:       rolesGrantPermissions.RoleId,
			PermissionId: permissionId,
			CreatedAt:    time.Now(),
		})
	}
	gormDB, _ := initializer.GetGorm()
	storeRoleWithPermissions := gormDB.Create(&roleWithPermissions)
	if storeRoleWithPermissions.Error != nil {
		message := fmt.Sprintf("failed to grant permissions for role id %d", rolesGrantPermissions.RoleId)
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   storeRoleWithPermissions.Error.Error(),
			"message": message,
		})
		context.Abort()
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    "grant role id (" + strconv.Itoa(int(rolesGrantPermissions.RoleId)) + ") with " + strconv.Itoa(int(storeRoleWithPermissions.RowsAffected)) + " permissions",
	})
}
