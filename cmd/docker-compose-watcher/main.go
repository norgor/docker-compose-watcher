package main

import (
	"docker-compose-watcher/internal/business"
)

func main() {
	c, err := business.NewComposeController("test.yml")
	defer c.Close()
	if err != nil {
		panic(err)
	}
	c.Run()
}
