package main

import (
	"errors"
	"fmt"
	"net/http"

	"jobbe.service/internal/data"
	"jobbe.service/internal/validator"
)

func (app *application) createSubscriberHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID int64  `json:"userID"`
		Tag    string `json:"tag"`
	}

	err := app.readJSON(w, r, &input)

	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	subscriber := &data.Subscriber{
		UserID: input.UserID,
		Tag:    input.Tag,
	}

	v := validator.New()

	if data.ValidateSubscriber(v, subscriber); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Subscribers.Insert(subscriber)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/subscribers/%d", subscriber.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"subscriber": subscriber}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteSubscriberHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Subscribers.Delete(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "subscriber successfully deleted"}, nil)
	if err != nil {

		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listSubscribersHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)

	if err != nil || id < 1 {
		app.notFoundResponse(w, r)
		return
	}

	subscribers, err := app.models.Subscribers.GetAllById(id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"subscribes": subscribers}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
