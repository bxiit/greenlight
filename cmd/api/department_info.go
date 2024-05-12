package main

import (
	"errors"
	"fmt"
	"github.com/bxiit/greenlight/internal/data"
	"net/http"
)

func (app *application) createDepInfoHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DepartmentName     string `json:"departmentName"`
		StaffQuantity      int    `json:"staffQuantity"`
		DepartmentDirector string `json:"departmentDirector"`
		ModuleId           int    `json:"moduleId"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, err.Error())
	}

	departmentInfo := &data.DepartmentInfo{
		DepartmentName:     input.DepartmentName,
		StaffQuantity:      input.StaffQuantity,
		DepartmentDirector: input.DepartmentDirector,
		ModuleId:           input.ModuleId,
	}

	err = app.models.DepartmentInfos.Insert(departmentInfo)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/department-info/%d", departmentInfo.ID))

	err = app.writeJSON(w, http.StatusCreated, Envelope{"department_info": departmentInfo}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getDepartmentInfoHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	departmentInfo, err := app.models.DepartmentInfos.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, Envelope{"department_info": departmentInfo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
