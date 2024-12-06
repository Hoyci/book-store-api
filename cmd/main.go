package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hoyci/book-store-api/cmd/api"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/db"
)

func main() {
	db := db.NewPGStorage()

	path := fmt.Sprintf("127.0.0.1:%s", config.Envs.Port)
	apiServer := api.NewApiServer(path, db)

	log.Println("Listening on:", path)
	http.ListenAndServe(path, apiServer)
}

// TODO: Corrigir error de last insert no book repository ✅
// TODO: Adicionar repository dinâmico para dar update no book
// TODO: Adicionar testes faltando para o book service e controller
