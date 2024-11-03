package handlers

import (
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PermissionsHandler struct {
}

func (p *PermissionsHandler) CreatePermissions(context *gin.Context) {
	var permissions struct {
		Permissions []models.Permission `json:"permissions" binding:"required"`
	}

	if errBind := context.ShouldBindJSON(&permissions); errBind != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "double check your JSON fields, kids",
		})
		context.Abort()
		return
	}

	gormDB, _ := initializer.GetGorm()
	createPermissions := gormDB.Create(&permissions.Permissions)
	if createPermissions.Error != nil {
		/* use switch for error database handling */
		context.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   createPermissions.Error.Error(),
			"message": "failed to store multiple permissions",
		})
		context.Abort()
		return
	}

	context.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    strconv.Itoa(int(createPermissions.RowsAffected)) + " permissions inserted successfully",
	})
}
