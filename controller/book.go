package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/hoyci/book-store-api/repository"
	"github.com/hoyci/book-store-api/service"
	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
)

var validate = validator.New()

type BookController struct {
	Service service.BookServiceInterface
}

func NewBookController(service service.BookServiceInterface) *BookController {
	return &BookController{
		Service: service,
	}
}

func (c *BookController) HandleCreateBook(w http.ResponseWriter, r *http.Request) {
	var payload types.CreateBookPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid input"))
		return
	}

	if err := validate.Struct(payload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("%v", errorMessages))
		return
	}

	id, err := c.Service.Create(r.Context(), payload)
	if err != nil {
		var insertErr *repository.InsertError
		var idErr *repository.LastInsertIDError

		if errors.As(err, &insertErr) {
			utils.WriteError(w, http.StatusInternalServerError, insertErr)
			return
		}

		if errors.As(err, &idErr) {
			utils.WriteError(w, http.StatusInternalServerError, idErr)
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("unexpected error: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (c *BookController) HandleGetBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("book ID must be a positive integer"))
		return
	}

	book, err := c.Service.GetByID(r.Context(), id)
	if err != nil {
		var notFoundErr *repository.ResourceNotFoundError
		var dbErr *repository.InternalDatabaseError

		if errors.As(err, &notFoundErr) {
			utils.WriteError(w, http.StatusNotFound, notFoundErr)
			return
		}

		if errors.As(err, &dbErr) {
			utils.WriteError(w, http.StatusInternalServerError, dbErr)
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("unexpected error: %w", err))
		return
	}

	utils.WriteJSON(w, http.StatusOK, book)
}

func (c *BookController) HandleUpdateBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, "book ID must be a positive integer")
		return
	}

	var payload types.UpdateBookPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid input")
		return
	}

	if err := validate.Struct(payload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, errorMessages)
		return
	}

	if payload.Name == nil && payload.Description == nil && payload.Author == nil &&
		payload.Genres == nil && payload.ReleaseYear == nil && payload.NumberOfPages == nil && payload.ImageUrl == nil {
		utils.WriteError(w, http.StatusBadRequest, "no fields provided for update")
		return
	}

	utils.WriteJSON(w, http.StatusOK, types.Book{
		ID:            1,
		Name:          "Hello World",
		Description:   "A programming book",
		Author:        "John Doe",
		Genres:        []string{"Fiction", "Technology"},
		ReleaseYear:   2024,
		NumberOfPages: 343,
		ImageUrl:      "http://example.com/go.jpg",
		CreatedAt:     time.Now(),
		UpdatedAt:     nil,
		DeletedAt:     nil,
	})
}

func (c *BookController) HandleDeleteBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, "book ID must be a positive integer")
		return
	}

	// returned_id, err := c.Service.DeleteByID(r.Context(), id)
	// if err != nil {
	// 	var notFoundErr *repository.ResourceNotFoundError
	// 	var dbErr *repository.InternalDatabaseError

	// 	if errors.As(err, &notFoundErr) {
	// 		utils.WriteError(w, http.StatusNotFound, notFoundErr)
	// 		return
	// 	}

	// 	if errors.As(err, &dbErr) {
	// 		utils.WriteError(w, http.StatusInternalServerError, dbErr)
	// 		return
	// 	}

	// 	utils.WriteError(w, http.StatusInternalServerError, fmt.Sprintf("unexpected error: %s", err.Error()))
	// 	return
	// }

	utils.WriteJSON(w, http.StatusOK, map[string]int{"id": 1})
}
