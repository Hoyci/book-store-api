package book

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
)

var validate = validator.New()

func updateBookPayloadStructLevelValidation(sl validator.StructLevel) {
	payload := sl.Current().Interface().(types.UpdateBookPayload)

	if payload.Name == nil &&
		payload.Description == nil &&
		payload.Author == nil &&
		payload.Genres == nil &&
		payload.ReleaseYear == nil &&
		payload.NumberOfPages == nil &&
		payload.ImageUrl == nil {
		sl.ReportError(payload, "UpdateBookPayload", "", "atleastonefield", "")
	}
}

type BookHandler struct {
	bookStore types.BookStore
}

func NewBookHandler(bookStore types.BookStore) *BookHandler {
	validate.RegisterStructValidation(updateBookPayloadStructLevelValidation, types.UpdateBookPayload{})

	return &BookHandler{bookStore: bookStore}
}

func (h *BookHandler) HandleCreateBook(w http.ResponseWriter, r *http.Request) {
	var payload types.CreateBookPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleCreateBook", "User sent request with an invalid JSON", "Body is not a valid json")
		return
	}

	if err := validate.Struct(payload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleCreateBook", "User sent a request containing JSON with information outside the permitted format", errorMessages)
		return
	}

	id, err := h.bookStore.Create(r.Context(), payload)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleCreateBook", "Failed to insert book into database", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]int{"id": id})
}

func (h *BookHandler) HandleGetBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleGetBookByID", "User sent request with an invalid ID", "Book ID must be a positive integer")
		return
	}

	book, err := h.bookStore.GetByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": fmt.Sprintf("No book found with ID %d", id)})
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBookByID", "Failed to get user by id from database", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusOK, book)
}

func (h *BookHandler) HandleUpdateBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateBookByID", "User sent request with an invalid ID", "Book ID must be a positive integer")
		return
	}

	var payload types.UpdateBookPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateBookByID", "User sent request with an invalid JSON", "Body is not a valid json")
		return
	}

	if err := validate.Struct(payload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateBookByID", "User sent a request with JSON outside the permitted format", errorMessages)
		return
	}

	book, err := h.bookStore.UpdateByID(r.Context(), id, payload)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": fmt.Sprintf("No book found with ID %d", id)})
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUpdateBookByID", "Failed to update book by id in database", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusOK, book)
}

func (h *BookHandler) HandleDeleteBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleDeleteBookByID", "User sent request with an invalid ID", "Book ID must be a positive integer")
		return
	}

	returnedID, err := h.bookStore.DeleteByID(r.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteJSON(w, http.StatusNotFound, map[string]string{"error": fmt.Sprintf("No book found with ID %d", id)})
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleDeleteBookByID", "Failed to delete user by id from database", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]int{"id": returnedID})
}
