package middlewares

import (
	initializer "go-no-cache-service/init"
	"go-no-cache-service/internal/helpers"
	"go-no-cache-service/internal/models"
	"log"
	"math"
	"net/http"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogs() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		measureHeader := ctx.GetHeader("X-Measure-Cache-Request-Logs")
		if measureHeader != "no-cache" {
			log.Println("no measurement requests...")
			ctx.Next()
			return
		}

		cacheSessionHeader := ctx.GetHeader("X-Cache-Session")
		cacheSessionID, errConv := strconv.Atoi(cacheSessionHeader)
		if errConv != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   errConv.Error(),
				"message": "measure cache request required X-Cache-Session Header",
			})

			ctx.Abort()
			return
		}

		debug.SetGCPercent(-1) // stopping automatic garbage collection

		cpuTimeBefore, errCpuTimeBefore := helpers.GetCPUStatUsage()
		if errCpuTimeBefore != nil {
			log.Println("ERROR CAPTURING CPU USAGE BEFORE!!!")
			log.Println(errCpuTimeBefore.Error())
		}

		startTimeRequest := time.Now()

		ctx.Next()

		responseTime := time.Since(startTimeRequest)

		memStats, errMemStats := helpers.CalcMemUsage()
		if errMemStats != nil {
			log.Println("ERROR MEM USAGE!!!")
			log.Println(errMemStats.Error())
		}

		cpuTimeAfter, errCpuTimeAfter := helpers.GetCPUStatUsage()
		if errCpuTimeAfter != nil {
			log.Println("ERROR CAPTURING CPU USAGE AFTER!!!")
			log.Println(errCpuTimeAfter.Error())
		}

		cpuStats, errCpuStats := helpers.CalcCPUTimeUsage(responseTime, (cpuTimeAfter - cpuTimeBefore))
		if errCpuStats != nil {
			log.Println(errCpuStats.Error())
		}

		debug.SetGCPercent(100) // starting automatic garbage collection

		var cacheHit, cacheMiss any
		var existHit, existMiss bool

		cacheHit, existHit = ctx.Get("CACHE_HIT")
		if !existHit {
			cacheHit = 0
		}
		cacheMiss, existMiss = ctx.Get("CACHE_MISS")
		if !existMiss {
			cacheMiss = 0
		}
		cacheType, existType := ctx.Get("CACHE_TYPE")
		if !existType {
			cacheType = "unset-cache-type"
		}

		gormDB, errGorm := initializer.GetMssqlDB()
		if errGorm != nil {
			panic(errGorm)
		}

		memoryUsageFormatted := math.Round(float64(memStats.Usage)/float64(1000_000)*100) / 100
		cpuUsageFormatted := math.Round(cpuStats.TotalCPUTime*100) / 100
		resourceUtilizationFormatted := math.Round(((memStats.Percent+cpuStats.TotalCPUTime)/2)*100) / 100

		m_requestLogs := models.RequestLog{
			CacheHit:            cacheHit.(int),
			CacheMiss:           cacheMiss.(int),
			ResponseTime:        uint64(responseTime.Milliseconds()),
			MemoryUsage:         memoryUsageFormatted,
			CPUUsage:            cpuUsageFormatted,
			ResourceUtilization: resourceUtilizationFormatted,
			CacheType:           cacheType.(string),
			CreatedAt:           time.Now(),
			CacheSessionID:      cacheSessionID,
		}
		errStoreLogs := gormDB.Create(&m_requestLogs).Error
		if errStoreLogs != nil {
			panic(errStoreLogs)
		}
	}
}
