package handlers

import (
	initializer "future-interns-backend/init"
	"future-interns-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PublicHandlers struct {
}

func (p *PublicHandlers) GetSkills(ctx *gin.Context) {
	var skills []map[string]interface{}
	gormDB, _ := initializer.GetGorm()

	gormDB.Model(&models.Skill{}).
		Select([]string{
			"skills.id",
			"skills.name",
			"images.id AS skill_icon_image_id",
		}).
		Joins("INNER JOIN images ON images.id = skills.skill_icon_image_id").
		Find(&skills)

	TransformsIdToPath([]string{"skill_icon_image_id"}, skills)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    skills,
	})
}

func (p *PublicHandlers) GetSocials(ctx *gin.Context) {
	var socials []map[string]interface{}
	gormDB, _ := initializer.GetGorm()

	gormDB.Model(&models.Social{}).Select([]string{"id", "name", "icon_image_id"}).Find(&socials)
	TransformsIdToPath([]string{"icon_image_id"}, socials)

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    socials,
	})
}
