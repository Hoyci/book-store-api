package user

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/hoyci/book-store-api/types"
	"github.com/hoyci/book-store-api/utils"
)

var validate = validator.New()

func PasswordValidator(sl validator.StructLevel) {
	data := sl.Current().Interface().(types.CreateUserRequestPayload)
	if data.Password != data.ConfirmPassword {
		sl.ReportError(data.ConfirmPassword, "ConfirmPassword", "ConfirmPassword", "password_mismatch", "")
	}
}

type UserHandler struct {
	userStore types.UserStore
}

func NewUserHandler(userStore types.UserStore) *UserHandler {
	validate.RegisterStructValidation(PasswordValidator, types.CreateUserRequestPayload{})

	return &UserHandler{userStore: userStore}
}

// Should not be a invalid/empty JSON ✅
// Should not miss (username, email, password and confirm_password) ✅
// Should be password and confirm_password equals
// Should be password and confirm_password lenght >= chars
// Should return a JWT valid token
func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var requestPayload types.CreateUserRequestPayload
	if err := utils.ParseJSON(r, &requestPayload); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "body is not a valid json")
		return
	}

	if err := validate.Struct(requestPayload); err != nil {
		var errorMessages []string
		for _, e := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, fmt.Sprintf("Field '%s' is invalid: %s", e.Field(), e.Tag()))
		}

		utils.WriteError(w, http.StatusBadRequest, errorMessages)
		return
	}

	hashedPassword, err := utils.HashPassword(requestPayload.Password)

	if err != nil {
		fmt.Println("An error occured during the hash password process")
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
	}

	var databasePayload = types.CreateUserDatabasePayload{
		Username:     requestPayload.Username,
		Email:        requestPayload.Email,
		PasswordHash: hashedPassword,
	}

	newUser, err := h.userStore.Create(r.Context(), databasePayload)
	if err != nil {
		utils.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	token, err := utils.CreateJWT(newUser.ID, newUser.Username, newUser.Email, "ABRACADABRA")
	if err != nil {
		fmt.Println("An error occured during the create JWT process")
		utils.WriteError(w, http.StatusInternalServerError, "internal server error")
	}

	utils.WriteJSON(w, http.StatusCreated, map[string]string{"token": token})
}
