package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/bxiit/greenlight/internal/data"
	"net/http"
	"time"
)

type MIHandler interface {
	CreateModuleInfoHandler(w http.ResponseWriter, r *http.Request)
	GetModuleInfoHandler(w http.ResponseWriter, r *http.Request)
	GetLatestFiftyModuleInfosHandler(w http.ResponseWriter, r *http.Request)
	EditModuleInfoHandler(w http.ResponseWriter, r *http.Request)
	DeleteModuleInfoHandler(w http.ResponseWriter, r *http.Request)
}

type ModuleInfoHandler struct {
	Repo data.ModuleInfoRepository
}

// new
func (mi *ModuleInfoHandler) CreateModuleInfoHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ModuleName     string        `json:"moduleName"`
		ModuleDuration time.Duration `json:"moduleDuration"`
		ExamType       string        `json:"examType"`
	}

	err := ReadJSON(w, r, &input)
	if err != nil {
		ErrorResponse(w, r, http.StatusBadRequest, err.Error())
	}

	moduleInfo := &data.ModuleInfo{
		ModuleName:     input.ModuleName,
		ModuleDuration: input.ModuleDuration,
		ExamType:       input.ExamType,
	}

	err = mi.Repo.Create(moduleInfo)
	if err != nil {
		ServerErrorResponse(w, r, err)
		return
	}

	//App.gormDB.Table("module_info").Create(&moduleInfo)
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/module-info/%d", moduleInfo.ID))

	//err = App.writeJSON(w, http.StatusCreated, Envelope{"module_info": moduleInfo}, headers)
	err = WriteJSON(w, http.StatusCreated, Envelope{"module_info": moduleInfo}, headers)
	if err != nil {
		ServerErrorResponse(w, r, err)
	}
}

func (mi *ModuleInfoHandler) GetModuleInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := ReadIDParam(r)
	if err != nil {
		NotFoundResponse(w, r)
	}

	//moduleInfo, err := App.models.ModuleInfos.Get(id)
	moduleInfo, err := mi.Repo.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			NotFoundResponse(w, r)
		default:
			ServerErrorResponse(w, r, err)
		}
		return
	}

	err = WriteJSON(w, http.StatusOK, Envelope{"module_info": moduleInfo}, nil)
	if err != nil {
		ServerErrorResponse(w, r, err)
	}
}

func (mi *ModuleInfoHandler) GetLatestFiftyModuleInfosHandler(w http.ResponseWriter, r *http.Request) {
	moduleInfos, err := mi.Repo.GetLatestFifty()

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			NotFoundResponse(w, r)
		}
	}

	err = WriteJSON(w, http.StatusOK, Envelope{"module_infos": moduleInfos}, nil)
	if err != nil {
		return
	}
}

func (mi *ModuleInfoHandler) EditModuleInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := ReadIDParam(r)
	if err != nil {
		NotFoundResponse(w, r)
	}

	moduleInfo, err := mi.Repo.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			NotFoundResponse(w, r)
		default:
			ServerErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		ModuleName     string        `json:"moduleName"`
		ModuleDuration time.Duration `json:"moduleDuration"`
		ExamType       string        `json:"examType"`
	}

	err = ReadJSON(w, r, &input)
	if err != nil {
		ServerErrorResponse(w, r, err)
		return
	}

	moduleInfo.ModuleName = input.ModuleName
	moduleInfo.ModuleDuration = input.ModuleDuration
	moduleInfo.ExamType = input.ExamType

	err = mi.Repo.Update(moduleInfo)
	if err != nil {
		ServerErrorResponse(w, r, err)
		return
	}

	err = WriteJSON(w, http.StatusOK, Envelope{"module_info": moduleInfo}, nil)
	if err != nil {
		ServerErrorResponse(w, r, err)
	}
}

func (mi *ModuleInfoHandler) DeleteModuleInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := ReadIDParam(r)
	if err != nil {
		NotFoundResponse(w, r)
	}

	err = mi.Repo.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			NotFoundResponse(w, r)
		default:
			ServerErrorResponse(w, r, err)
		}
		return
	}

	err = WriteJSON(w, http.StatusOK, Envelope{"message": "module info successfully deleted"}, nil)
	if err != nil {
		ServerErrorResponse(w, r, err)
	}
}

// old
func (app *application) CreateModuleInfoHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ModuleName     string        `json:"moduleName"`
		ModuleDuration time.Duration `json:"moduleDuration"`
		ExamType       string        `json:"examType"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
	}

	moduleInfo := &data.ModuleInfo{
		ModuleName:     input.ModuleName,
		ModuleDuration: input.ModuleDuration,
		ExamType:       input.ExamType,
	}

	err = app.models.ModuleInfos.Create(moduleInfo)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	//App.gormDB.Table("module_info").Create(&moduleInfo)
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/module-info/%d", moduleInfo.ID))

	err = app.writeJSON(w, http.StatusCreated, Envelope{"module_info": moduleInfo}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) GetModuleInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	moduleInfo, err := app.models.ModuleInfos.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, Envelope{"module_info": moduleInfo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) GetLatestFiftyModuleInfosHandler(w http.ResponseWriter, r *http.Request) {
	moduleInfos, err := app.models.ModuleInfos.GetLatestFifty()

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.notFoundResponse(w, r)
		}
	}

	err = app.writeJSON(w, http.StatusOK, Envelope{"module_infos": moduleInfos}, nil)
	if err != nil {
		return
	}
}

func (app *application) EditModuleInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	moduleInfo, err := app.models.ModuleInfos.Get(id)
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
		ModuleName     string        `json:"moduleName"`
		ModuleDuration time.Duration `json:"moduleDuration"`
		ExamType       string        `json:"examType"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	moduleInfo.ModuleName = input.ModuleName
	moduleInfo.ModuleDuration = input.ModuleDuration
	moduleInfo.ExamType = input.ExamType

	err = app.models.ModuleInfos.Update(moduleInfo)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, Envelope{"module_info": moduleInfo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) DeleteModuleInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	err = app.models.ModuleInfos.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, Envelope{"message": "module info successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
