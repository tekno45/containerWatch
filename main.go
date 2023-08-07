package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	//keep getting containers to check for labels.
	for {
		containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
		if err != nil {
			panic(err)
		}
		//check for our label indicating a container to watch
		watchContainers(containers)
	}
}

func checkThread(container types.Container, checkRequest http.Request, ctx context.Context) {
	fmt.Println("Launching Thread to check: ", container.Names[0])
	pollingTime, ok := ctx.Value("pollingTime").(int)
	if !ok {
		panic("Polling Time not found")
	}

	for {
		select {
		case <-ctx.Done():
			return

		case <-time.After(time.Duration(pollingTime) * time.Second):
			req, err := http.DefaultClient.Do(&checkRequest)
			fmt.Println("Check: ", req.StatusCode)
			if err != nil {
				fmt.Println(err)
				return
			}

		}

	}

}

func launchThread(container *types.Container) {
	req, err := http.NewRequest("GET", "http://"+container.Labels["healthWatch_url"], nil)
	go checkThread(*container, *req, context.WithValue(context.Background(), "pollingTime", 5))
	if err != nil {
		panic(err)
	}
}

func watchContainers(containers []types.Container) {
	launchedIDs := make(map[string]int)
	for _, container := range containers {
		_, ok := container.Labels["healthWatch"]
		if ok {
			//Check if container already is being watched, or at least was launched
			_, ok := launchedIDs[container.ID]
			if ok {
				continue
			}
			launchedIDs[container.ID] = 1
			launchThread(&container)
			continue
		}
		time.Sleep(time.Second * 2)
	}
}
