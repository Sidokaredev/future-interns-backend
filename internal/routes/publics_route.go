package routes

import (
	"future-interns-backend/internal/constants"
	"future-interns-backend/internal/handlers"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Address struct {
	Street string `json:"street" cache:"street"`
	Code   int    `json:"code" cache:"code"`
}

type Session struct {
	SignedAt  time.Time `json:"signed_at" cache:"signed_at"`
	ExpiredAt time.Time `json:"expired_at" cache:"expired_at"`
}

type CacheData struct {
	Id       any     `json:"id" cache:"id"`
	Position string  `json:"position" cache:"position"`
	Salary   int     `json:"salary" cache:"salary"`
	Address  Address `json:"address" cache:"address"`
	Session  Session `json:"session" cache:"session"`
}

func PublicRoutes(apiv1 *gin.RouterGroup) {
	publicHandlers := &handlers.PublicHandlers{}

	routePublic := apiv1.Group("/public")

	routeSkill := routePublic.Group("/skills")
	{
		routeSkill.Handle(http.MethodGet, "/", publicHandlers.GetSkills)
	}
	routeSocial := routePublic.Group("/socials")
	{
		routeSocial.Handle(constants.MethodGet, "/", publicHandlers.GetSocials)
	}

	// routeCache := routePublic.Group("/cache")
	// {
	// 	routeCache.Handle("POST", "/", func(ctx *gin.Context) {
	// 		var body CacheData
	// 		if errBind := ctx.ShouldBindJSON(&body); errBind != nil {
	// 			ctx.JSON(http.StatusBadRequest, gin.H{
	// 				"success": false,
	// 				"error":   errBind.Error(),
	// 				"message": "gagal melakukan binding JSON pada permintaan",
	// 			})
	// 			return
	// 		}

	// 		// should able to check whether the body is a single Map or Struct or Slice
	// 		rds_hash := caching.ExtractToHash("id", []any{body})
	// 		rds_sorted_set := caching.NewSortedSetCollection(rds_hash, caching.SortedSetArgs{
	// 			ScorePropName:  "session.signed_at",
	// 			MemberPropName: "id",
	// 			ScoreType:      "time.Time",
	// 		})

	// 		log.Println(rds_sorted_set)

	// 		ctx.JSON(http.StatusOK, gin.H{
	// 			"success": true,
	// 			"data":    rds_sorted_set,
	// 		})
	// 	})
	// }
}
