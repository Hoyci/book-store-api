package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
	"go.opentelemetry.io/otel"
)

var validate = validator.New()

func passwordValidator(sl validator.StructLevel) {
	data := sl.Current().Interface().(types.CreateUserRequestPayload)
	if data.Password != data.ConfirmPassword {
		sl.ReportError(data.ConfirmPassword, "ConfirmPassword", "ConfirmPassword", "password_mismatch", "")
	}
}

type UserHandler struct {
	userStore types.UserStore
}

func NewUserHandler(userStore types.UserStore) *UserHandler {
	validate.RegisterStructValidation(passwordValidator, types.CreateUserRequestPayload{})

	return &UserHandler{userStore: userStore}
}

// @Summary Criar um novo usuário
// @Tags Users
// @Accept json
// @Produce json
// @Param request body types.CreateUserRequestPayload true "Payload contendo os dados do novo usuário"
// @Success 201 {object} types.CreateUserResponse "Usuário criado com sucesso"
// @Failure 400 {object} types.BadRequestResponse "Body is not a valid json"
// @Failure 400 {object} types.BadRequestStructResponse "Validation errors for payload"
// @Failure 500 {object} types.InternalServerErrorResponse "An unexpected error occurred"
// @Failure 503 {object} types.ContextCanceledResponse "Request canceled"
// @Router /users [post]
func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := otel.Tracer("user-handler")

	_, validationSpan := tracer.Start(ctx, "ValidateRequest")
	var requestPayload types.CreateUserRequestPayload
	if err := utils.ParseJSON(r, &requestPayload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleCreateUser", types.BadRequestResponse{Error: "Body is not a valid json"})
		return
	}

	if err := validate.Struct(requestPayload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleCreateUser", types.BadRequestStructResponse{Error: errorMessages})
		return
	}
	validationSpan.End()

	user, _ := h.userStore.GetByEmail(r.Context(), requestPayload.Email)

	if user != nil {
		utils.WriteError(w, http.StatusConflict, fmt.Errorf("email is already in use"), "HandleCreateUser", types.InternalServerErrorResponse{Error: "This email is already in use"})
		return
	}

	hashedPassword, err := utils.HashPassword(ctx, requestPayload.Password)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleCreateUser", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
	}

	var databasePayload = types.CreateUserDatabasePayload{
		Username:     requestPayload.Username,
		Email:        requestPayload.Email,
		PasswordHash: hashedPassword,
	}

	_, err = h.userStore.Create(r.Context(), databasePayload)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleCreateUser", types.ContextCanceledResponse{Error: "Request canceled"})
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleCreateUser", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
			return
		}

		utils.WriteError(w, http.StatusInternalServerError, err, "HandleCreateUser", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	utils.WriteJSON(w, http.StatusCreated, types.CreateUserResponse{Message: "User successfully created"})
}

// @Summary      Get user by ID
// @Description  Retrieves user details based on the authenticated user's ID extracted from the request context.
// @Tags         Users
// @Security BearerAuth
// @Accept       json
// @Produce      json
// @Success      200  {object}  types.UserResponse  "User details successfully retrieved"
// @Failure      400  {object}  types.BadRequestResponse "Bad request"
// @Failure      401  {object}  types.UnauthorizedResponse "Unauthorized"
// @Failure      404  {object}  types.NotFoundResponse "User not found"
// @Failure      500  {object}  types.InternalServerErrorResponse "Internal server error"
// @Router       /users [get]
func (h *UserHandler) HandleGetUserByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetClaimFromContext[int](r, "UserID")
	if !ok {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to retrieve userID from context"), "HandleGetUserByID", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	user, err := h.userStore.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleGetUserByID", types.ContextCanceledResponse{Error: "Request canceled"})
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetUserByID", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
			return
		}

		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleGetUserByID", types.NotFoundResponse{Error: fmt.Sprintf("No user found with ID %d", userID)})
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleGetUserByID", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

// @Summary      Update user by ID
// @Description  Updates user details based on the authenticated user's ID extracted from the request context.
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param request body  types.UpdateUserPayload  true  "User update payload"
// @Success      200  {object}  types.UserResponse  "User successfully updated"
// @Failure      400  {object}  types.BadRequestResponse "Invalid request body"
// @Failure      401  {object}  types.UnauthorizedResponse "Unauthorized"
// @Failure      404  {object}  types.NotFoundResponse "User not found"
// @Failure      500  {object}  types.InternalServerErrorResponse "Internal server error"
// @Router       /users [put]
func (h *UserHandler) HandleUpdateUserByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetClaimFromContext[int](r, "UserID")
	if !ok {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to retrieve userID from context"), "HandleUpdateUserByID", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	var payload types.UpdateUserPayload
	if err := utils.ParseJSON(r, &payload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateUserByID", types.BadRequestResponse{Error: "Body is not a valid json"})
		return
	}

	if err := validate.Struct(payload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, err, "HandleUpdateUserByID", types.BadRequestStructResponse{Error: errorMessages})
		return
	}

	user, err := h.userStore.UpdateByID(r.Context(), userID, payload)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleUpdateUserByID", types.ContextCanceledResponse{Error: "Request canceled"})
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleUpdateUserByID", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
			return
		}

		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleUpdateUserByID", types.NotFoundResponse{Error: fmt.Sprintf("No user found with ID %d", userID)})
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleUpdateUserByID", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	utils.WriteJSON(w, http.StatusOK, user)
}

// @Summary      Delete user by ID
// @Description  Deletes the user associated with the authenticated user's ID extracted from the request context.
// @Tags         Users
// @Security     BearerAuth
// @Success      204  "User successfully deleted"
// @Failure      401  {object}  types.UnauthorizedResponse "Unauthorized"
// @Failure      404  {object}  types.NotFoundResponse "User not found"
// @Failure      500  {object}  types.InternalServerErrorResponse "Internal server error"
// @Router       /users [delete]
func (h *UserHandler) HandleDeleteUserByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetClaimFromContext[int](r, "UserID")
	if !ok {
		utils.WriteError(w, http.StatusInternalServerError, fmt.Errorf("failed to retrieve userID from context"), "HandleDeleteUserByID", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	err := h.userStore.DeleteByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			utils.WriteError(w, http.StatusServiceUnavailable, err, "HandleDeleteUserByID", types.ContextCanceledResponse{Error: "Request canceled"})
			return
		}

		if err == sql.ErrConnDone {
			utils.WriteError(w, http.StatusInternalServerError, err, "HandleDeleteUserByID", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
			return
		}

		if err == sql.ErrNoRows {
			utils.WriteError(w, http.StatusNotFound, err, "HandleDeleteUserByID", types.NotFoundResponse{Error: fmt.Sprintf("No user found with ID %d", userID)})
			return
		}
		utils.WriteError(w, http.StatusInternalServerError, err, "HandleDeleteUserByID", types.InternalServerErrorResponse{Error: "An unexpected error occurred"})
		return
	}

	utils.WriteJSON(w, http.StatusNoContent, nil)
}
