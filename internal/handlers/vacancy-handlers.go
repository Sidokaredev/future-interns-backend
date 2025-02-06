package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type VacancyHandler struct {
}

func (handler *VacancyHandler) GetVacanciesReadThrough(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    "read through service successfully",
	})
}
