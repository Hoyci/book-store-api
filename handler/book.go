package handler

import (
	"net/http"

	"github.com/hoyci/book-store-api/service"
)

type BookHandler struct {
	Service *service.BookService
}

func NewBookHandler(service *service.BookService) *BookHandler {
	return &BookHandler{
		Service: service,
	}
}

func (h *BookHandler) HandleCreateBook(w http.ResponseWriter, r *http.Request) {
	// TODO: Tratar os dados que vem para verificar se estão corretos
	// response, err := h.Service.Create()
}

func (h *BookHandler) HandleGetBookByID(w http.ResponseWriter, r *http.Request) {
	// TODO: Checar se o ID que está vindo na URL é um número posito inteiro
	// response, err := h.Service.GetByID()
}
