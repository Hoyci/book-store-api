package main

import (
	"fmt"
	"log"

	"github.com/hoyci/book-store-api/cmd/api"
	"github.com/hoyci/book-store-api/config"
)

func main() {
	server := api.NewApiServer(fmt.Sprintf("127.0.0.1:%s", config.Envs.Port))
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
