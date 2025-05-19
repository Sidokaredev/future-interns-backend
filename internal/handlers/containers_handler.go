package handlers

import (
	"context"
	"fmt"
	initializer "future-interns-backend/init"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/gin-gonic/gin"
)

type ContainersHandler struct {
}

func (c *ContainersHandler) ListContainers(ctx *gin.Context) {
	cli, errCli := initializer.GetDockerClient()
	if errCli != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCli.Error(),
			"message": "gagal memanggil docker client",
		})

		return
	}

	contxt := context.Background()
	containers, errContainers := cli.ContainerList(contxt, container.ListOptions{All: true})
	if errContainers != nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   errContainers.Error(),
			"message": "gagal mendapatkan daftar container",
		})

		return
	}

	type ContainerServicesList struct {
		ContainerID   string
		ContainerName string
		Status        string
	}
	var listContainerServices []ContainerServicesList

	for _, container := range containers {
		for _, name := range container.Names {
			containerName := strings.TrimPrefix(name, "/")
			if strings.HasPrefix(containerName, "skripsi") {
				listContainerServices = append(listContainerServices, ContainerServicesList{
					ContainerID:   container.ID[:12],
					ContainerName: containerName,
					Status:        container.Status,
				})
			}
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    listContainerServices,
	})
}

func (c *ContainersHandler) ContainersControl(ctx *gin.Context) {
	type DockerCommand struct {
		Action      string `json:"action" binding:"required"`
		ContainerID string `json:"container_id" binding:"required"`
	}

	var command DockerCommand

	if errBind := ctx.ShouldBindJSON(&command); errBind != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errBind.Error(),
			"message": "invalid json fields",
		})

		return
	}

	contxt := context.Background()
	cli, errCli := initializer.GetDockerClient()
	if errCli != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   errCli.Error(),
			"message": "gagal memanggil docker client",
		})

		return
	}

	switch command.Action {
	case "start":
		errContainerStart := cli.ContainerStart(contxt, command.ContainerID, container.StartOptions{})
		if errContainerStart != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errContainerStart.Error(),
				"message": fmt.Sprintf("gagal menjalankan container ID: %s", command.ContainerID),
			})

			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"message": fmt.Sprintf("berhasil menjalankan container ID: %s", command.ContainerID),
			},
		})
		return

	case "stop":
		errContainerStop := cli.ContainerStop(contxt, command.ContainerID, container.StopOptions{})
		if errContainerStop != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   errContainerStop.Error(),
				"message": fmt.Sprintf("gagal menghentikan container ID: %s", command.ContainerID),
			})

			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"message": fmt.Sprintf("berhasil menghentikan container ID: %s", command.ContainerID),
			},
		})
		return

	default:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "tidak ada aksi yang sesuai",
			"message": "aksi container: [start, stop]",
		})

		return
	}
}
