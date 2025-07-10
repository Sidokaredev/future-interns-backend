package writer_pipelines

import (
	"context"
	"fmt"
	initializer "go-write-behind-service/init"
	"go-write-behind-service/internal/models"
	"go-write-behind-service/internal/services/caches"
	"log"
	"reflect"
	"time"
)

func PipelinesWriteScheduler() error {
	rdb, errRedis := initializer.GetRedisDB()
	if errRedis != nil {
		return errRedis
	}

	ctx := context.Background()
	const count int = 100

	// -> mengambil pipelineID dari antrean List [job:writer@pipelines]
	pipelines, errListPipelines := rdb.RPopCount(ctx, caches.JobPipelinesKey, count).Result()
	if errListPipelines != nil {
		return errListPipelines
	}

	pipelinesToWrite := []map[string]any{}
	sortedSetKeysCheck := map[string]bool{}
	sortedSetKeys := []string{}

	for _, key := range pipelines {
		// -> mengambil semua field-value Hash by pipelineID
		hash, errScanning := rdb.HGetAll(ctx, key).Result()
		if errScanning != nil {
			return errScanning
		}

		// ERR: harusnya candidate_id
		if !sortedSetKeysCheck[hash["candidate_id"]] {
			sortedSetKeysCheck[hash["candidate_id"]] = true
			sortedSetKeys = append(sortedSetKeys, hash["candidate_id"])
		}

		pipelineMapAny := make(map[string]any)

		// -> perulangan untuk mengubah map[string]string ke map[string]any
		iteration := reflect.ValueOf(hash).MapRange()
		for iteration.Next() {
			key := iteration.Key().String()
			// -> mengonversi nilai string [created_at] RFC3339 ke bentuk time.Time
			if key == "created_at" {
				timeCreatedAt, errParse := time.Parse(time.RFC3339, hash["created_at"])
				if errParse != nil {
					return errParse
				}
				pipelineMapAny[key] = timeCreatedAt
				continue
			}

			val := iteration.Value().Interface()
			pipelineMapAny[key] = val
		}

		pipelinesToWrite = append(pipelinesToWrite, pipelineMapAny)
	}

	DB, errDB := initializer.GetMssqlDB()
	if errDB != nil {
		return errDB
	}

	// -> melakukan write/insert 100 data dengan create in batches 10
	/*
		NOTE:
		jika insert gagal, cache kembali nilai yang dikembalikan oleh RPUSH
	*/
	errPipelines := DB.Model(&models.Pipeline{}).CreateInBatches(pipelinesToWrite, 10).Error
	if errPipelines != nil {
		return errPipelines
	}

	// -> memberikan/mengatur ulang expiration time sortedset pemilik pipelines
	/*
	   NOTE:
	   jika Sorted Set telah memiliki TTL dan masing-masing Hash juga memiliki TTL. Ketika Pipeline baru ditambahkan, TTL Sorted Set akan diatur ulang dan Hash dari Pipeline baru memiliki expiration time baru, akan tetapi Hash dari Pipeline yang lainnya tidak diperbarui.
	*/
	pipe := rdb.Pipeline()
	for _, key := range sortedSetKeys {
		prefix := fmt.Sprintf("pipelines:%v", key)
		pipe.Expire(ctx, prefix, 1*time.Hour)
	}

	if cmds, errExec := pipe.Exec(ctx); errExec != nil {
		for _, cmd := range cmds {
			log.Printf("cmd: %v | args: %v | err: %v", cmd.FullName(), cmd.Args(), cmd.Err())
		}
		return errExec
	}

	// -> memberikan expiration time untuk pipeline Hash
	hash := caches.ExtractToHash("id", pipelinesToWrite)
	errHSet := hash.SetExpireCollection(1 * time.Hour)
	if errHSet != nil {
		return errHSet
	}

	return nil
}
