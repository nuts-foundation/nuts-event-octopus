// Package api provides primitives to interact the openapi HTTP API.
//
// This is an autogenerated file, any edits which you make here will be lost!
package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
	"net/http"
	"strings"
)

// Event defines component schema for Event.
type Event struct {
	ConsentId  string     `json:"consentId"`
	Custodian  Identifier `json:"custodian"`
	Error      *string    `json:"error,omitempty"`
	ExternalId string     `json:"externalId"`
	Payload    string     `json:"payload"`
	RetryCount int32      `json:"retryCount"`
	State      string     `json:"state"`
	Uuid       string     `json:"uuid"`
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
	// Find a specific event (POST /events/{uuid})
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
	router.POST("/events/:uuid", wrapper.GetEvent)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/7RV74/cNBD9V0YGiVbKZXd7p1LlG7RQLVTX6koRUjmps84ka3Ds1B7vNVT7v6Nxsr/u",
	"FhU+cB9OXmfGfjPvvfFnpX3Xe0eOo6o+q6jX1GFe/rAhx7Log+8psKG8rb2L5HhZy4+aog6mZ+OdqtSv",
	"V/Du3fIFYIymdVTDaoDnPtQI7AEhkPahVoVqfOiQVaVSMvKbh55UpSIH41q1LZROkX1t0MkdXwdqVKW+",
	"mh2QziaYs2VNjk1jKEgaheDDQ1h5GwJh9A6MA42RwDeA0CSnJQgtjLlnsNAnpuDQnit4+QI0Wp0s8lit",
	"DkPPHlI0roXv314Duhr6YDbIBH/SINceijtzW4+D9Xjmqmu6ez52/oY+Jor8luXMn96+vgaMgFpTvwMx",
	"xl2sgqlbgkdVoKb64BLHi9NPF9ibD4/PwQjEYXju0yiAUyRzYfO3Q5ZxTO1IQBRMkkEudap6r8KIlYRl",
	"3zQU8oo9rGgPWRXqaNkYh9b8dRTXU4hmOkMkYGlcj4zdnkGfZfVP6vyy/nL9H5MRsNX7XdBY20lrTsRR",
	"HDnjWMEHTm+3xWiqVybyDcVe4h8ajDY7MxqmLn7JAqNNt/sqMAQc1FaqOHLHg268JEfBaDD7GEiRamiy",
	"VfpAUsqk4gKwXWlfUwHEuoQlfxMB7R0OMYuNQ9KiPdGhg3c319B4a/3dKEeE2qeVJdDeegePqsfZFbwm",
	"J/92CAa5bIM2ZW/Kh9ZsKB/3u8utRuFeWAuu8qaunpSLp+Wzq3m5KBeLy2fPLssn5VX5tFxU8/Hv2zPc",
	"bgtlXOMfNuS7N0uIPWnTGI2yl1uRyYDIPlAJv6zpeGMaaDGDzfLYITfuorGmXfMYHstcgTWaJsYddoLq",
	"5ZtXm8vMneFc2nXieHKFIFKF2ogHRveVi3IuKb4nh71Rlbos5+U864zXWS6zg4Rayg4WfeWiZIopayJn",
	"JY8SzIFP5vNpuvM09rHv7dSK2R/Ru8Pz8K8keSLy3PfTfr/+GXYAClj5eoC1t3UEASd9nGqQxJi6DsOg",
	"KnVDnIIDtHb6DjqFQI7tIJM990xIxzaKdXOMupUzpp7MPoudt9l1Pp7pTUs8Okr6GbAjpiBn3ZeLHJOn",
	"+Wh6mCYdYH5QCmjFX7tnoTEhTlpQoj9VZbJUsRPCNGMOY4dDouKo3V8aWbf/N53/hUKfWPsuu2Fqiwj2",
	"an51DxLTJ571Fk0GczD4aADnGRqfXB6+e6Bn/HzvodylFdAHv8KVHaBOJI9WoMgYeBo/KYLzYL1rKUxW",
	"prqAJvgO7tZGr0Gjyw9VjKkjSUEGw7DGCCsiB/vXqLwn0x+NqwH342TP/ANdbrd/BwAA///Z7qMSgQkA",
	"AA==",
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file.
func GetSwagger() (*openapi3.Swagger, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error loading Swagger: %s", err)
	}
	return swagger, nil
}
