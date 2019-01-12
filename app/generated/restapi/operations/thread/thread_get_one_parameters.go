// Code generated by go-swagger; DO NOT EDIT.

package thread

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime/middleware"
	strfmt "github.com/go-openapi/strfmt"
)

// NewThreadGetOneParams creates a new ThreadGetOneParams object
// no default values defined in spec.
func NewThreadGetOneParams() ThreadGetOneParams {

	return ThreadGetOneParams{}
}

// ThreadGetOneParams contains all the bound params for the thread get one operation
// typically these are obtained from a http.Request
//
// swagger:parameters threadGetOne
type ThreadGetOneParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*Идентификатор ветки обсуждения.
	  Required: true
	  In: path
	*/
	SlugOrID string
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewThreadGetOneParams() beforehand.
func (o *ThreadGetOneParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	rSlugOrID, rhkSlugOrID, _ := route.Params.GetOK("slug_or_id")
	if err := o.bindSlugOrID(rSlugOrID, rhkSlugOrID, route.Formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindSlugOrID binds and validates parameter SlugOrID from path.
func (o *ThreadGetOneParams) bindSlugOrID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route

	o.SlugOrID = raw

	return nil
}
