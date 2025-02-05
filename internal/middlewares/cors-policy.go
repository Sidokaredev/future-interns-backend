package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func CORSPolicy() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// var mStart runtime.MemStats
		// runtime.ReadMemStats(&mStart)

		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")

		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}

		ctx.Next()

		// var mEnd runtime.MemStats
		// runtime.ReadMemStats(&mEnd)

		// memUsedMB := float64(mEnd.Alloc - mStart.Alloc)
		// rdb, err := initializer.GetRedisDB()
		// if err != nil {
		// 	log.Println("err redis database")
		// 	ctx.Abort()
		// 	return
		// }

		// c := context.Background()
		// errSetKey := rdb.Set(c, "memused", memUsedMB, 0).Err()
		// if errSetKey != nil {
		// 	log.Println("failed to set memused")
		// 	ctx.Abort()

		// 	return
		// }
	}
}
