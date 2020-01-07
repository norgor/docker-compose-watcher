package main

import (
	"docker-compose-watcher/internal/serviceprovider"
	"fmt"
	"io"
)

func main() {
	provider, err := serviceprovider.NewServiceProvider()
	if err != nil {
		panic(err)
	}
	defer provider.Close()

	for {
		ebs := <-provider.Channel
		if ebs.Error != nil {
			if ebs.Error == io.EOF {
				continue
			}
			panic(ebs.Error)
		}
		for k, v := range ebs.Services {
			fmt.Printf("%s: %v\n", k, v)
		}
	}
}
