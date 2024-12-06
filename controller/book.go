package controller

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

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
		errors := err.(validator.ValidationErrors)

		for _, e := range errors {
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

	utils.WriteJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func (c *BookController) HandleGetBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid book ID: must be a positive integer"))
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

func (c *BookController) HandleDeleteBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, fmt.Errorf("invalid book ID: must be a positive integer"))
		return
	}

	err = c.Service.DeleteByID(r.Context(), id)
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

}
