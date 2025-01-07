package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hoyci/book-store-api/cmd/api"
	"github.com/hoyci/book-store-api/config"
	"github.com/hoyci/book-store-api/db"
	"github.com/hoyci/book-store-api/service/book"
	"github.com/hoyci/book-store-api/service/healthcheck"
	"github.com/hoyci/book-store-api/service/user"
)

func main() {
	db := db.NewPGStorage()
	path := fmt.Sprintf("127.0.0.1:%s", config.Envs.Port)

	apiServer := api.NewApiServer(path, db)

	healthCheckHandler := healthcheck.NewHealthCheckHandler(config.Envs)

	bookStore := book.NewBookStore(db)
	bookHandler := book.NewBookHandler(bookStore)

	userStore := user.NewUserStore(db)
	userHandler := user.NewUserHandler(userStore)

	apiServer.SetupRouter(healthCheckHandler, bookHandler, userHandler)

	log.Println("Listening on:", path)
	http.ListenAndServe(path, apiServer.Router)
}

// TODO: Adicionar endpoint de refresh token
// TODO: Adicionar swagger para documentar a API
// TODO: Adicionar restrição nos endpoints de usuários e books (somente o proprio usuário pode alterar e deletar suas informações) / (somente o proprio usuário pode alterar e deletar informações dos seus livros)
// TODO: Deve ser possível que o usuário atribua um livro a ele (Criar um projeto tipo o Skoob)
