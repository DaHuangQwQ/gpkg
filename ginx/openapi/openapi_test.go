package openapi

import (
	"testing"
)

type req struct {
	Name string `json:"name,omitempty" validate:"required,min=1"`
	Age  int    `json:"age,omitempty" validate:"required"`
}

type res struct {
	Code int `json:"code"`
}

func TestRegisterOpenAPIOperation(t *testing.T) {

}
