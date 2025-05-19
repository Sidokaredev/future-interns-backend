package initializer

import (
	"errors"
	"log"
	"sync"

	"github.com/docker/docker/client"
)

var (
	dockercli  *client.Client
	onceDocker sync.Once
)

func DockerClientInit() {
	onceDocker.Do(func() {
		var err error
		dockercli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			log.Println("docker client fail to connect")
			panic(err)
		}

		log.Printf("Connected to Docker: %v", dockercli)
	})
}

func GetDockerClient() (*client.Client, error) {
	if dockercli == nil {
		return nil, errors.New("docker client hasn't started")
	}

	return dockercli, nil
}
