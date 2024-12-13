package book

import (
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

func (h *BookHandler) HandleCreateBook(w http.ResponseWriter, r *http.Request) {
	var payload types.CreateBookPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "body is not a valid json")
		return
	}

	if err := validate.Struct(payload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, errorMessages)
		return
	}

	id, err := h.bookStore.Create(r.Context(), payload)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (h *BookHandler) HandleGetBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, "book ID must be a positive integer")
		return
	}

	book, err := h.bookStore.GetByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, book)
}

func (h *BookHandler) HandleUpdateBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, "book ID must be a positive integer")
		return
	}

	var payload types.UpdateBookPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "body is not a valid json")
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

	book, err := h.bookStore.UpdateByID(r.Context(), id, payload)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, book)
}

func (h *BookHandler) HandleDeleteBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, "book ID must be a positive integer")
		return
	}

	returnedID, err := h.bookStore.DeleteByID(r.Context(), id)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]int{"id": returnedID})
}
