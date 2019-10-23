// Package api provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package api

import (
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/labstack/echo/v4"
	"net/http"
)

// Event defines component schema for Event.
type Event struct {
	ConsentId            *string    `json:"consentId,omitempty"`
	Error                *string    `json:"error,omitempty"`
	ExternalId           string     `json:"externalId"`
	InitiatorLegalEntity Identifier `json:"initiatorLegalEntity"`
	Name                 string     `json:"name"`
	Payload              string     `json:"payload"`
	RetryCount           int32      `json:"retryCount"`
	TransactionId        *string    `json:"transactionId,omitempty"`
	Uuid                 string     `json:"uuid"`
}

// EventListResponse defines component schema for EventListResponse.
type EventListResponse struct {
	Events []Event `json:"events,omitempty"`
}

// Identifier defines component schema for Identifier.
type Identifier string

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Return all events currently in store (GET /events)
	List(ctx echo.Context) error
	// Find a specific event by its externalId (GET /events/by_external_id/{external_id})
	GetEventByExternalId(ctx echo.Context, externalId string) error
	// Find a specific event (GET /events/{uuid})
	GetEvent(ctx echo.Context, uuid string) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// List converts echo context to params.
func (w *ServerInterfaceWrapper) List(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.List(ctx)
	return err
}

// GetEventByExternalID converts echo context to params.
func (w *ServerInterfaceWrapper) GetEventByExternalId(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "external_id" -------------
	var externalId string

	err = runtime.BindStyledParameter("simple", false, "external_id", ctx.Param("external_id"), &externalId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter external_id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetEventByExternalId(ctx, externalId)
	return err
}

// GetEvent converts echo context to params.
func (w *ServerInterfaceWrapper) GetEvent(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "uuid" -------------
	var uuid string

	err = runtime.BindStyledParameter("simple", false, "uuid", ctx.Param("uuid"), &uuid)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter uuid: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetEvent(ctx, uuid)
	return err
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router runtime.EchoRouter, si ServerInterface) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET("/events", wrapper.List)
	router.GET("/events/by_external_id/:external_id", wrapper.GetEventByExternalId)
	router.GET("/events/:uuid", wrapper.GetEvent)

}
