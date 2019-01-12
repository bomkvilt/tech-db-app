// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// ThreadUpdate Сообщение для обновления ветки обсуждения на форуме.
// Пустые параметры остаются без изменений.
//
// swagger:model ThreadUpdate
type ThreadUpdate struct {

	// Описание ветки обсуждения.
	Message string `json:"message,omitempty"`

	// Заголовок ветки обсуждения.
	Title string `json:"title,omitempty"`
}

// Validate validates this thread update
func (m *ThreadUpdate) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *ThreadUpdate) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *ThreadUpdate) UnmarshalBinary(b []byte) error {
	var res ThreadUpdate
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}