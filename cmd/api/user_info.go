package main

import (
	"context"
	"errors"
	"github.com/bxiit/greenlight/internal/data"
	"github.com/bxiit/greenlight/internal/validator"
	pb "github.com/bxiit/greenlight/rpc"
	"net/http"
	"time"
)

type UserService struct {
	pb.UnimplementedUserServiceServer
}

type UIHandler interface {
	ActivateUserInfoHandler(w http.ResponseWriter, r *http.Request)
	RegisterUserInfoHandler(w http.ResponseWriter, r *http.Request)
	GetUserInfoHandler(w http.ResponseWriter, r *http.Request)
	GetAllUserInfoHandler(w http.ResponseWriter, r *http.Request)
	EditUserInfoHandler(w http.ResponseWriter, r *http.Request)
	DeleteUserInfoHandler(w http.ResponseWriter, r *http.Request)
}

type UserInfoHandler struct {
	Repo                 data.UserInfoRepository
	TokenRepoForUI       data.TokenRepo
	PermissionsRepoForUI data.PermissionRepo
	app                  *application
}

func (c *UserService) InsertUser(ctx context.Context, req *pb.UserRequest) (*pb.UserResponse, error) {
	return &pb.UserResponse{Ok: req.Email == "atabekbekseiit@gmail.com"}, nil
}

func (app *application) ActivateUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the plaintext activation token from the request body.
	var input struct {
		TokenPlaintext string `json:"token"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// Validate the plaintext token provided by the client.
	v := validator.New()
	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Retrieve the details of the user associated with the token using the
	// GetForToken() method (which we will create in a minute). If no matching record
	// is found, then we let the client know that the token they provided is not valid.

	userInfo, err := app.models.UserInfos.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Update the user's activation status.

	userInfo.Activated = true
	// Save the updated user record in our database, checking for any edit conflicts in
	// the same way that we did for our movie records.

	err = app.models.UserInfos.Update(userInfo)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// If everything went successfully, then we delete all activation tokens for the
	// user.
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, userInfo.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	// Send the updated user details to the client in a JSON response.
	err = app.writeJSON(w, http.StatusOK, Envelope{"userInfo": userInfo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) RegisterUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Surname  string `json:"surname"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	userInfo := &data.UserInfo{
		Name:      input.Name,
		Surname:   input.Surname,
		Email:     input.Email,
		Activated: false,
		Role:      "user",
	}

	err = userInfo.PasswordHashed.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUserInfo(v, userInfo); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.UserInfos.Insert(userInfo)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.models.Permissions.AddForUser(userInfo.ID, "movies:read")
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	token, err := app.models.Tokens.New(userInfo.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userInfoID":      userInfo.ID,
		}
		err := app.mailer.Send(userInfo.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	err = app.writeJSON(w, http.StatusAccepted, Envelope{"userInfo": userInfo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) GetUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	userInfoCtx := app.contextGetUserInfo(r)

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	if userInfoCtx.ID != id {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	userInfo, err := app.models.UserInfos.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, Envelope{"user_info": userInfo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) GetAllUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	userInfos, err := app.models.UserInfos.GetAll()
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, Envelope{"user_infos": userInfos}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) EditUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	userInfo, err := app.models.UserInfos.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Name    string `json:"name"`
		Surname string `json:"surname"`
		Email   string `json:"email"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	userInfo.Name = input.Name
	userInfo.Surname = input.Surname
	userInfo.Email = input.Email

	err = app.models.UserInfos.Update(userInfo)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, Envelope{"userInfo": userInfo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) DeleteUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	err = app.models.UserInfos.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, Envelope{"message": "user info successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (ui *UserInfoHandler) ActivateUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the plaintext activation token from the request body.
	var input struct {
		TokenPlaintext string `json:"token"`
	}
	err := ReadJSON(w, r, &input)
	if err != nil {
		BadRequestResponse(w, r, err)
		return
	}
	// Validate the plaintext token provided by the client.
	v := validator.New()
	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		FailedValidationResponse(w, r, v.Errors)
		return
	}
	// Retrieve the details of the user associated with the token using the
	// GetForToken() method (which we will create in a minute). If no matching record
	// is found, then we let the client know that the token they provided is not valid.

	userInfo, err := ui.Repo.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			FailedValidationResponse(w, r, v.Errors)
		default:
			ServerErrorResponse(w, r, err)
		}
		return
	}
	// Update the user's activation status.

	userInfo.Activated = true
	// Save the updated user record in our database, checking for any edit conflicts in
	// the same way that we did for our movie records.

	err = ui.Repo.Update(userInfo)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			EditConflictResponse(w, r)
		default:
			ServerErrorResponse(w, r, err)
		}
		return
	}
	// If everything went successfully, then we delete all activation tokens for the
	// user.
	err = ui.TokenRepoForUI.DeleteAllForUser(data.ScopeActivation, userInfo.ID)
	if err != nil {
		ServerErrorResponse(w, r, err)
		return
	}
	// Send the updated user details to the client in a JSON response.
	err = WriteJSON(w, http.StatusOK, Envelope{"userInfo": userInfo}, nil)
	if err != nil {
		ServerErrorResponse(w, r, err)
	}
}

func (ui *UserInfoHandler) RegisterUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Surname  string `json:"surname"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := ReadJSON(w, r, &input)
	if err != nil {
		BadRequestResponse(w, r, err)
		return
	}

	userInfo := &data.UserInfo{
		Name:      input.Name,
		Surname:   input.Surname,
		Email:     input.Email,
		Activated: false,
		Role:      "user",
	}

	err = userInfo.PasswordHashed.Set(input.Password)
	if err != nil {
		ServerErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUserInfo(v, userInfo); !v.Valid() {
		FailedValidationResponse(w, r, v.Errors)
		return
	}

	err = ui.Repo.Insert(userInfo)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			FailedValidationResponse(w, r, v.Errors)
		default:
			ServerErrorResponse(w, r, err)
		}
		return
	}

	err = ui.PermissionsRepoForUI.AddForUser(userInfo.ID, "movies:read")
	if err != nil {
		ServerErrorResponse(w, r, err)
		return
	}

	token, err := ui.TokenRepoForUI.New(userInfo.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		ServerErrorResponse(w, r, err)
		return
	}

	ui.app.background(func() {
		data := map[string]any{
			"activationToken": token.Plaintext,
			"userInfoID":      userInfo.ID,
		}
		err := ui.app.mailer.Send(userInfo.Email, "user_welcome.tmpl", data)
		if err != nil {
			ui.app.logger.PrintError(err, nil)
		}
	})

	err = WriteJSON(w, http.StatusAccepted, Envelope{"userInfo": userInfo}, nil)
	if err != nil {
		ServerErrorResponse(w, r, err)
	}
}
