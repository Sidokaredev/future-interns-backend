package middlewares

import (
	"context"
	"fmt"
	initializer "go-read-through-service/init"
	"go-read-through-service/internal/models"
	"log"
	"math"
	"net/http"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/cpu"
)

var totalRequest atomic.Int64

func RequestLogs() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		measureHeader := ctx.GetHeader("X-Measure-Cache-Request-Logs")
		if measureHeader != "cache-aside" {
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

		startTimeRequest := time.Now() // capturing starting time of request

		var memStart runtime.MemStats
		runtime.ReadMemStats(&memStart) // read memory start when request begin

		ctx.Next()

		if ctx.Writer.Status() != http.StatusOK {
			ctx.Abort() // Hentikan eksekusi jika bukan 200
			return
		}

		cpuUsage, _ := cpu.Percent(0, false)

		responseTime := time.Since(startTimeRequest) // capturing response time

		var memEnd runtime.MemStats
		runtime.ReadMemStats(&memEnd) // read memory stats after request processed

		debug.SetGCPercent(100) // starting automatic garbage collection

		memUsed := float64((memEnd.Alloc - memStart.Alloc) / (1024 * 1024)) // calculating memory usage per-request in MB
		memoryUtilization := float64(memUsed/float64(512)) * 100            // calculate memory usage percentage from allocated machine

		cacheHit, _ := ctx.Get("CACHE_HIT")
		cacheMiss, _ := ctx.Get("CACHE_MISS")

		rdb, errRedis := initializer.GetRedisDB()
		if errRedis != nil {
			panic(errRedis)
		}
		c := context.Background()
		errHSet := rdb.HSet(c, "request:"+strconv.Itoa(int(totalRequest.Load())), map[string]interface{}{
			"cacheHit":            cacheHit,
			"cacheMiss":           cacheMiss,
			"responseTime":        responseTime.Milliseconds(),
			"memoryUsage":         math.Round((memoryUtilization * 100)) / 100,
			"cpuUsage":            math.Round((cpuUsage[0] * 100)) / 100,
			"resourceUtilization": math.Round((cpuUsage[0]*0.5)+(memoryUtilization*0.5)*100) / 100,
		}).Err()
		if errHSet != nil {
			panic(errHSet)
		}

		totalRequest.Add(1) // count request

		if totalRequest.Load() == 20 {
			cursor := uint64(0)
			m_requestLogs := []models.RequestLog{}

			for {
				var requestKeys []string
				var err error
				requestKeys, cursor, err = rdb.Scan(c, cursor, "request:*", 0).Result()
				if err != nil {
					log.Fatalf("failed scan 'request:*' \t: %v", err)
				}

				for _, key := range requestKeys {
					var HCursor uint64
					for {
						var requestLogs map[string]string
						requestLogs, err = rdb.HGetAll(c, key).Result()
						if err != nil {
							log.Fatalf("failed HSCAN %s: %v", key, err)
						}

						logs := map[string]interface{}{}
						for keyIndex, logValue := range requestLogs {
							if keyIndex == "cacheHit" || keyIndex == "cacheMiss" {
								cacheValue, errCache := strconv.Atoi(logValue)
								if errCache != nil {
									log.Fatal(errCache)
								}

								logs[keyIndex] = cacheValue
								continue
							}
							if keyIndex == "responseTime" {
								cacheValue, errCache := strconv.ParseUint(logValue, 10, 64)
								if errCache != nil {
									log.Fatal(errCache)
								}

								logs[keyIndex] = cacheValue
								continue
							}
							if keyIndex == "memoryUsage" || keyIndex == "cpuUsage" || keyIndex == "resourceUtilization" {
								cacheValue, errCache := strconv.ParseFloat(logValue, 64)
								if errCache != nil {
									log.Fatal(errCache)
								}

								logs[keyIndex] = cacheValue
								continue
							}
						}
						m_requestLogs = append(m_requestLogs, models.RequestLog{
							CacheHit:            logs["cacheHit"].(int),
							CacheMiss:           logs["cacheMiss"].(int),
							ResponseTime:        logs["responseTime"].(uint64),
							MemoryUsage:         logs["memoryUsage"].(float64),
							CPUUsage:            logs["cpuUsage"].(float64),
							ResourceUtilization: logs["resourceUtilization"].(float64),
							CacheType:           "cache-aside",
							CreatedAt:           time.Now(),
							CacheSessionID:      cacheSessionID,
						})
						if HCursor == 0 {
							break
						}
					}
				}

				if len(requestKeys) > 0 {
					_, err := rdb.Del(c, requestKeys...).Result()
					if err != nil {
						log.Fatalf("fail deleting 'request:*' keys: %v", err)
					}
					fmt.Println("'request:*' keys deleted")
				}

				if cursor == 0 {
					break
				}
			}

			gormDB, errGorm := initializer.GetMssqlDB()
			if errGorm != nil {
				log.Fatal(errGorm)
			}
			errStoreReqeustLogs := gormDB.CreateInBatches(&m_requestLogs, 10).Error
			if errStoreReqeustLogs != nil {
				log.Fatal(errStoreReqeustLogs)
			}

			totalRequest.Store(0) // reset counter after 500 requests
		}

		log.Println("cache-hit status \t:", cacheHit.(bool))
		log.Println("cache-miss status \t:", cacheMiss.(bool))

		log.Println("response time \t:", responseTime.Milliseconds())

		log.Printf("mem usage /request \t: %vMB", strconv.FormatFloat(memUsed, 'f', 2, 64))
		log.Printf("percent mem usage /request \t: %v%%", strconv.FormatFloat(memoryUtilization, 'f', 2, 64))
		log.Printf("Resource Utilization \t: %v%%", math.Round((cpuUsage[0]*0.5)+(memoryUtilization*0.5)*100)/100)
	}
}
