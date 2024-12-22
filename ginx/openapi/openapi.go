package openapi

import (
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
)

func NewOpenAPI() *OpenAPI {
	desc := NewOpenApiSpec()
	return &OpenAPI{
		description:            &desc,
		generator:              openapi3gen.NewGenerator(),
		globalOpenAPIResponses: []openAPIResponse{},
	}
}

// OpenAPI Holds the OpenAPI OpenAPIDescription (OAD) and OpenAPI capabilities.
type OpenAPI struct {
	description            *openapi3.T
	generator              *openapi3gen.Generator
	globalOpenAPIResponses []openAPIResponse
}

func (openAPI *OpenAPI) Description() *openapi3.T {
	return openAPI.description
}

func (openAPI *OpenAPI) Generator() *openapi3gen.Generator {
	return openAPI.generator
}

// Compute the tags to declare at the root of the OpenAPI spec from the tags declared in the operations.
func (openAPI *OpenAPI) computeTags() {
	for _, pathItem := range openAPI.Description().Paths.Map() {
		for _, op := range pathItem.Operations() {
			for _, tag := range op.Tags {
				if openAPI.Description().Tags.Get(tag) == nil {
					openAPI.Description().Tags = append(openAPI.Description().Tags, &openapi3.Tag{
						Name: tag,
					})
				}
			}
		}
	}

	// Make sure tags are sorted
	slices.SortFunc(openAPI.Description().Tags, func(a, b *openapi3.Tag) int {
		return strings.Compare(a.Name, b.Name)
	})
}

// getOrCreateSchema is used to get a schema from the OpenAPI spec.
// If the schema does not exist, it will create a new schema and add it to the OpenAPI spec.
func (openAPI *OpenAPI) getOrCreateSchema(key string, v any) *openapi3.Schema {
	schemaRef, ok := openAPI.Description().Components.Schemas[key]
	if !ok {
		schemaRef = openAPI.createSchema(key, v)
	}
	return schemaRef.Value
}

// createSchema is used to create a new schema and add it to the OpenAPI spec.
// Relies on the openapi3gen package to generate the schema, and adds custom struct tags.
func (openAPI *OpenAPI) createSchema(key string, v any) *openapi3.SchemaRef {
	schemaRef, err := openAPI.Generator().NewSchemaRefForValue(v, openAPI.Description().Components.Schemas)
	if err != nil {
		slog.Error("Error generating schema", "key", key, "error", err)
	}
	schemaRef.Value.Description = key + " schema"

	descriptionable, ok := v.(Descriptioner)
	if ok {
		schemaRef.Value.Description = descriptionable.Description()
	}

	parseStructTags(reflect.TypeOf(v), schemaRef)

	openAPI.Description().Components.Schemas[key] = schemaRef

	return schemaRef
}

func NewOpenApiSpec() openapi3.T {
	const openapiDescription = "123"
	info := &openapi3.Info{
		Title:       "OpenAPI",
		Description: openapiDescription,
		Version:     "0.0.1",
	}
	spec := openapi3.T{
		OpenAPI:  "3.1.0",
		Info:     info,
		Paths:    &openapi3.Paths{},
		Servers:  []*openapi3.Server{},
		Security: openapi3.SecurityRequirements{},
		Components: &openapi3.Components{
			Schemas:       make(map[string]*openapi3.SchemaRef),
			RequestBodies: make(map[string]*openapi3.RequestBodyRef),
			Responses:     make(map[string]*openapi3.ResponseRef),
		},
	}
	return spec
}

// DefaultDescription returns a default .md description for a controller
func DefaultDescription[T any](handler string, middlewares []T) string {
	description := "#### Controller: \n\n`" +
		handler + "`"

	if len(middlewares) > 0 {
		description += "\n\n#### Middlewares:\n"

		for i, fn := range middlewares {
			description += "\n- `" + FuncName(fn) + "`"

			if i == 4 {
				description += "\n- more middleware…"
				break
			}
		}
	}

	return description + "\n\n---\n\n"
}

func registerOpenAPIOperation[T, B any](openapi *OpenAPI, route Path[T, B]) (*openapi3.Operation, error) {
	if route.Operation == nil {
		route.Operation = openapi3.NewOperation()
	}

	if route.FullName == "" {
		route.FullName = route.Path
	}

	route.GenerateDefaultDescription()

	if route.Operation.Summary == "" {
		route.Operation.Summary = route.NameFromNamespace(camelToHuman)
	}

	if route.Operation.OperationID == "" {
		route.GenerateDefaultOperationID()
	}

	// Request Body
	if route.Operation.RequestBody == nil {
		bodyTag := SchemaTagFromType(openapi, *new(B))

		if bodyTag.Name != "unknown-interface" {
			requestBody := newRequestBody[B](bodyTag, route.AcceptedContentTypes)

			// add request body to operation
			route.Operation.RequestBody = &openapi3.RequestBodyRef{
				Value: requestBody,
			}
		}
	}

	// Response - globals
	for _, openAPIGlobalResponse := range openapi.globalOpenAPIResponses {
		addResponseIfNotSet(
			openapi,
			route.Operation,
			openAPIGlobalResponse.Code,
			openAPIGlobalResponse.Description,
			openAPIGlobalResponse.Response,
		)
	}

	// Automatically add non-declared 200 (or other) Response
	if route.DefaultStatusCode == 0 {
		route.DefaultStatusCode = 200
	}
	defaultStatusCode := strconv.Itoa(route.DefaultStatusCode)
	responseDefault := route.Operation.Responses.Value(defaultStatusCode)
	if responseDefault == nil {
		response := openapi3.NewResponse().WithDescription(http.StatusText(route.DefaultStatusCode))
		route.Operation.AddResponse(route.DefaultStatusCode, response)
		responseDefault = route.Operation.Responses.Value(defaultStatusCode)
	}

	// Automatically add non-declared Content for 200 (or other) Response
	if responseDefault.Value.Content == nil {
		responseSchema := SchemaTagFromType(openapi, *new(T))
		content := openapi3.NewContentWithSchemaRef(&responseSchema.SchemaRef, []string{"application/json", "application/xml"})
		responseDefault.Value.WithContent(content)
	}

	// Automatically add non-declared Path parameters
	for _, pathParam := range parsePathParams(route.Path) {
		if exists := route.Operation.Parameters.GetByInAndName("path", pathParam); exists != nil {
			continue
		}
		parameter := openapi3.NewPathParameter(pathParam)
		parameter.Schema = openapi3.NewStringSchema().NewRef()
		if strings.HasSuffix(pathParam, "...") {
			parameter.Description += " (might contain slashes)"
		}

		route.Operation.AddParameter(parameter)
	}
	for _, params := range route.Operation.Parameters {
		if params.Value.In == "path" {
			if !strings.Contains(route.Path, "{"+params.Value.Name) {
				panic(fmt.Errorf("path parameter '%s' is not declared in the path", params.Value.Name))
			}
		}
	}

	openapi.Description().AddOperation(route.Path, route.Method, route.Operation)

	return route.Operation, nil
}

func newRequestBody[RequestBody any](tag SchemaTag, consumes []string) *openapi3.RequestBody {
	content := openapi3.NewContentWithSchemaRef(&tag.SchemaRef, consumes)
	return openapi3.NewRequestBody().
		WithRequired(true).
		WithDescription("Request body for " + reflect.TypeOf(*new(RequestBody)).String()).
		WithContent(content)
}
