package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type VacancyHandlers struct {
}

func (h *VacancyHandlers) GetVacancies(context *gin.Context) {
	context.JSON(http.StatusOK, gin.H{
		"status": "middleware passed",
	})
}
