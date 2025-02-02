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

func (h *BookHandler) HandleCreateBook(w http.ResponseWriter, r *http.Request) {
	var payload types.CreateBookPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleCreateBook", "Body is not a valid json")
		return
	}

	if err := validate.Struct(payload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleCreateBook", errorMessages)
		return
	}

	id, err := h.bookStore.Create(r.Context(), payload)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleGetBooks", "Request canceled")
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBooks", "An unexpected error occurred")
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, err, "HandleCreateBook", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusCreated, types.CreateBookResponse{ID: id})
}

func (h *BookHandler) HandleGetBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleGetBookByID", "Book ID must be a positive integer")
		return
	}

	book, err := h.bookStore.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleGetBooks", "Request canceled")
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBooks", "An unexpected error occurred")
			return
		}

		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleGetBookByID", fmt.Sprintf("No book found with ID %d", id))
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBookByID", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusOK, book)
}

func (h *BookHandler) HandleGetBooks(w http.ResponseWriter, r *http.Request) {
	books, err := h.bookStore.GetMany(r.Context())
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleGetBooks", "Request canceled")
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBooks", "An unexpected error occurred")
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBooks", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string][]*types.Book{"books": books})
}

func (h *BookHandler) HandleUpdateBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateBookByID", "Book ID must be a positive integer")
		return
	}

	var payload types.UpdateBookPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateBookByID", "Body is not a valid json")
		return
	}

	if err := validate.Struct(payload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateBookByID", errorMessages)
		return
	}

	book, err := h.bookStore.UpdateByID(r.Context(), id, payload)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleGetBooks", "Request canceled")
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetBooks", "An unexpected error occurred")
			return
		}

		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleGetBookByID", fmt.Sprintf("No book found with ID %d", id))
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUpdateBookByID", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusOK, book)
}

func (h *BookHandler) HandleDeleteBookByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleDeleteBookByID", "Book ID must be a positive integer")
		return
	}

	err = h.bookStore.DeleteByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleDeleteBookByID", "Request canceled")
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleDeleteBookByID", "An unexpected error occurred")
			return
		}

		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleDeleteBookByID", fmt.Sprintf("No book found with ID %d", id))
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, err, "HandleDeleteBookByID", "An unexpected error occurred")
		return
	}

	utils.WriteJSON(w, http.StatusNoContent, nil)
}
