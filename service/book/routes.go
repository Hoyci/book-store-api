package book

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
)

var validate = validator.New()

type BookHandler struct {
	bookStore types.BookStore
}

func NewBookHandler(bookStore types.BookStore) *BookHandler {
	return &BookHandler{bookStore: bookStore}
}

// @Summary Criar novo livro
// @Tags Books
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body types.CreateBookPayload true "Dados do livro"
// @Success 201 {object} types.CreateBookResponse "ID do novo livro"
// @Failure 400 {object} types.BadRequestResponse "Body is not a valid json"
// @Failure 400 {object} types.BadRequestStructResponse "Validation errors for payload"
// @Failure 500 {object} types.InternalServerErrorResponse "An unexpected error occurred"
// @Failure 503 {object} types.ContextCanceledResponse "Request canceled"
// @Router /books [post]
func (h *BookHandler) HandleCreateBook(w http.ResponseWriter, r *http.Request) {
	var payload types.CreateBookPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleCreateBook", types.BadRequestResponse{Error: "Body is not a valid json"})
		return
	}

	if err := validate.Struct(payload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleCreateBook", types.BadRequestStructResponse{Error: errorMessages})
		return
	}

	id, err := h.bookStore.Create(r.Context(), payload)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleGetBooks", types.ContextCanceledResponse{Error: "Request canceled"})
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBooks", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, err, "HandleCreateBook", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, types.CreateBookResponse{ID: id})
}

// @Summary Obter livro por ID
// @Tags Books
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID do livro"
// @Success 200 {object} types.Book "Detalhes do livro"
// @Failure 400 {object} types.BadRequestResponse "Book ID must be a positive integer"
// @Failure 404 {object} types.NotFoundResponse "No book found with given ID"
// @Failure 500 {object} types.InternalServerErrorResponse "An unexpected error occurred"
// @Failure 503 {object} types.ContextCanceledResponse "Request canceled"
// @Router /books/{id} [get]
func (h *BookHandler) HandleGetBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleGetBookByID", types.BadRequestResponse{Error: "Book ID must be a positive integer"})
		return
	}

	book, err := h.bookStore.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleGetBookByID", types.ContextCanceledResponse{Error: "Request canceled"})
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBookByID", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
			return
		}

		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleGetBookByID", types.NotFoundResponse{Error: fmt.Sprintf("No book found with ID %d", id)})
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBookByID", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.Book{
		ID:            book.ID,
		Name:          book.Name,
		Description:   book.Description,
		Author:        book.Author,
		Genres:        book.Genres,
		ReleaseYear:   book.ReleaseYear,
		NumberOfPages: book.NumberOfPages,
		ImageUrl:      book.ImageUrl,
		CreatedAt:     book.CreatedAt,
		DeletedAt:     book.DeletedAt,
		UpdatedAt:     book.UpdatedAt,
	})
}

// @Summary Listar livros
// @Tags Books
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} types.GetBooksResponse "Lista de livros"
// @Failure 500 {object} types.InternalServerErrorResponse "An unexpected error occurred"
// @Failure 503 {object} types.ContextCanceledResponse "Request canceled"
// @Router /books [get]
func (h *BookHandler) HandleGetBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.bookStore.GetMany(r.Context())
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleGetBooks", types.ContextCanceledResponse{Error: "Request canceled"})
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBooks", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBooks", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.GetBooksResponse{Books: books})
}

// @Summary Atualizar livro por ID
// @Tags Books
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID do livro a ser atualizado"
// @Param request body types.UpdateBookPayload true "Dados para atualização do livro"
// @Success 200 {object} types.Book "Livro atualizado"
// @Failure 400 {object} types.BadRequestResponse "Book ID must be a positive integer ou Body is not a valid json"
// @Failure 400 {object} types.BadRequestStructResponse "Validation errors for payload"
// @Failure 404 {object} types.NotFoundResponse "No book found with given ID"
// @Failure 500 {object} types.InternalServerErrorResponse "An unexpected error occurred"
// @Failure 503 {object} types.ContextCanceledResponse "Request canceled"
// @Router /books/{id} [put]
func (h *BookHandler) HandleUpdateBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateBookByID", types.BadRequestResponse{Error: "Book ID must be a positive integer"})
		return
	}

	var payload types.UpdateBookPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateBookByID", types.BadRequestResponse{Error: "Body is not a valid json"})
		return
	}

	if err := validate.Struct(payload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateBookByID", types.BadRequestStructResponse{Error: errorMessages})
		return
	}

	book, err := h.bookStore.UpdateByID(r.Context(), id, payload)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleGetBooks", types.ContextCanceledResponse{Error: "Request canceled"})
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBooks", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
			return
		}

		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleGetBookByID", types.NotFoundResponse{Error: fmt.Sprintf("No book found with ID %d", id)})
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUpdateBookByID", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.Book{
		ID:            book.ID,
		Name:          book.Name,
		Description:   book.Description,
		Author:        book.Author,
		Genres:        book.Genres,
		ReleaseYear:   book.ReleaseYear,
		NumberOfPages: book.NumberOfPages,
		ImageUrl:      book.ImageUrl,
		CreatedAt:     book.CreatedAt,
		DeletedAt:     book.DeletedAt,
		UpdatedAt:     book.UpdatedAt,
	})
}

// @Summary Excluir livro por ID
// @Tags Books
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "ID do livro a ser excluído"
// @Success 204 "No Content"
// @Failure 400 {object} types.BadRequestResponse "Book ID must be a positive integer"
// @Failure 404 {object} types.NotFoundResponse "No book found with given ID"
// @Failure 500 {object} types.InternalServerErrorResponse "An unexpected error occurred"
// @Failure 503 {object} types.ContextCanceledResponse "Request canceled"
// @Router /books/{id} [delete]
func (h *BookHandler) HandleDeleteBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleDeleteBookByID", types.BadRequestResponse{Error: "Book ID must be a positive integer"})
		return
	}

	err = h.bookStore.DeleteByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleDeleteBookByID", types.ContextCanceledResponse{Error: "Request canceled"})
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleDeleteBookByID", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
			return
		}

		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleDeleteBookByID", types.NotFoundResponse{Error: fmt.Sprintf("No book found with ID %d", id)})
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, err, "HandleDeleteBookByID", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	utils.WriteJSON(w, http.StatusNoContent, nil)
}
