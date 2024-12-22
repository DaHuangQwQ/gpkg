package openapi

import (
	"strconv"

	"github.com/getkin/kin-openapi/openapi3"
)

type OpenAPIParam struct {
	Name        string
	Description string

	// Default value for the parameter.
	// Type is checked at start-time.
	Default  any
	Example  string
	Examples map[string]any
	Type     ParamType

	// integer, string, bool
	GoType string

	// Status codes for which this parameter is required.
	// Only used for response parameters.
	// If empty, it is required for 200 status codes.
	StatusCodes []int

	Required bool
	Nullable bool
}

type ParamType string // Query, Header, Cookie

const (
	PathParamType   ParamType = "path"
	QueryParamType  ParamType = "query"
	HeaderParamType ParamType = "header"
	CookieParamType ParamType = "cookie"
)

// Registers a response for the route, only if error for this code is not already set.
func addResponseIfNotSet(openapi *OpenAPI, operation *openapi3.Operation, code int, description string, response Response) {
	if operation.Responses.Value(strconv.Itoa(code)) != nil {
		return
	}
	operation.AddResponse(code, openapi.buildOpenapi3Response(description, response))
}

func (openAPI *OpenAPI) buildOpenapi3Response(description string, response Response) *openapi3.Response {
	if response.Type == nil {
		panic("Type in Response cannot be nil")
	}

	responseSchema := SchemaTagFromType(openAPI, response.Type)
	if len(response.ContentTypes) == 0 {
		response.ContentTypes = []string{"application/json", "application/xml"}
	}

	content := openapi3.NewContentWithSchemaRef(&responseSchema.SchemaRef, response.ContentTypes)
	return openapi3.NewResponse().
		WithDescription(description).
		WithContent(content)
}

// openAPIResponse describes a response error in the OpenAPI spec.
type openAPIResponse struct {
	Response
	Code        int
	Description string
}

// when setting custom response types on routes
type Response struct {
	// content-type of the response i.e application/json
	ContentTypes []string
	// user provided type
	Type any
}
