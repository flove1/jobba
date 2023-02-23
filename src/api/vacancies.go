package main

import (
	"errors"
	"fmt"
	"net/http"

	"jobbe.service/internal/data"
	"jobbe.service/internal/validator"
)

func (app *application) createVacancyHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string   `json:"title"`
		Company string   `json:"company"`
		Tags    []string `json:"tags"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	vacancy := &data.Vacancy{
		Title:   input.Title,
		Company: input.Company,
		Tags:    input.Tags,
	}

	v := validator.New()

	if data.ValidateVacancy(v, vacancy); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Vacancies.Insert(vacancy)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/vacancys/%d", vacancy.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"vacancy": vacancy}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	app.background(func() {
		data := map[string]any{
			"title":   vacancy.Title,
			"company": vacancy.Company,
			"tags":    vacancy.Tags,
		}
		for _, tag := range vacancy.Tags {
			subscribers, err := app.models.Subscribers.GetAllByTag(tag)
			if err != nil {
				app.serverErrorResponse(w, r, err)
			}
			for _, subscriber := range subscribers {
				err = app.mailer.Send(subscriber.Email, "new_vacancy.tmpl", data)
				if err != nil {
					app.logger.PrintError(err, nil)
				}
			}
		}
	})
}

func (app *application) deleteVacancyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Vacancies.Delete(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "vacancy successfully deleted"}, nil)
	if err != nil {

		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateVacancyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	vacancy, err := app.models.Vacancies.Get(id)
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
		Title   *string  `json:"title"`
		Company *string  `json:"company"`
		Tags    []string `json:"tags"`
		Active  *bool    `json:"active"`
	}

	err = app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		vacancy.Title = *input.Title
	}

	if input.Company != nil {
		vacancy.Company = *input.Company
	}

	if input.Tags != nil {
		vacancy.Tags = input.Tags
	}

	if input.Active != nil {
		vacancy.Active = *input.Active
	}

	v := validator.New()

	if data.ValidateVacancy(v, vacancy); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Vacancies.Update(vacancy)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"vacancy": vacancy}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) showVacancyHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	vacancy, err := app.models.Vacancies.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"vacancy": vacancy}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listVacanciesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string
		Tags  []string
		data.Filters
	}
	v := validator.New()
	qs := r.URL.Query()
	input.Title = app.readString(qs, "title", "")
	input.Tags = app.readCSV(qs, "tags", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "title", "company", "-id", "-title", "-company"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	vacancies, metadata, err := app.models.Vacancies.GetAll(input.Title, input.Tags, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"vacancies": vacancies, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
