package openapi

import (
	"github.com/getkin/kin-openapi/openapi3"
	"log/slog"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// SchemaTag is a struct that holds the name of the struct and the associated openapi3.SchemaRef
type SchemaTag struct {
	openapi3.SchemaRef
	Name string
}

// parseStructTags parses struct tags and modifies the schema accordingly.
// t must be a struct type.
// It adds the following struct tags (tag => OpenAPI schema field):
// - description => description
// - example => example
// - json => nullable (if contains omitempty)
// - validate:
//   - required => required
//   - min=1 => min=1 (for integers)
//   - min=1 => minLength=1 (for strings)
//   - max=100 => max=100 (for integers)
//   - max=100 => maxLength=100 (for strings)
func parseStructTags(t reflect.Type, schemaRef *openapi3.SchemaRef) {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}

	for i := range t.NumField() {
		field := t.Field(i)

		if field.Anonymous {
			fieldType := field.Type
			parseStructTags(fieldType, schemaRef)
			continue
		}

		jsonFieldName := field.Tag.Get("json")
		jsonFieldName = strings.Split(jsonFieldName, ",")[0] // remove omitempty, etc
		if jsonFieldName == "-" {
			continue
		}
		if jsonFieldName == "" {
			jsonFieldName = field.Name
		}

		property := schemaRef.Value.Properties[jsonFieldName]
		if property == nil {
			slog.Warn("Property not found in schema", "property", jsonFieldName)
			continue
		}
		propertyCopy := *property
		propertyValue := *propertyCopy.Value

		// Example
		example, ok := field.Tag.Lookup("example")
		if ok {
			propertyValue.Example = example
			if propertyValue.Type.Is(openapi3.TypeInteger) {
				exNum, err := strconv.Atoi(example)
				if err != nil {
					slog.Warn("Example might be incorrect (should be integer)", "error", err)
				}
				propertyValue.Example = exNum
			}
		}

		// Validation
		validateTag, ok := field.Tag.Lookup("validate")
		validateTags := strings.Split(validateTag, ",")
		if ok && slices.Contains(validateTags, "required") {
			schemaRef.Value.Required = append(schemaRef.Value.Required, jsonFieldName)
		}
		for _, validateTag := range validateTags {
			if strings.HasPrefix(validateTag, "min=") {
				min, err := strconv.Atoi(strings.Split(validateTag, "=")[1])
				if err != nil {
					slog.Warn("Min might be incorrect (should be integer)", "error", err)
				}

				if propertyValue.Type.Is(openapi3.TypeInteger) {
					minPtr := float64(min)
					propertyValue.Min = &minPtr
				} else if propertyValue.Type.Is(openapi3.TypeString) {
					//nolint:gosec // disable G115
					propertyValue.MinLength = uint64(min)
				}
			}
			if strings.HasPrefix(validateTag, "max=") {
				max, err := strconv.Atoi(strings.Split(validateTag, "=")[1])
				if err != nil {
					slog.Warn("Max might be incorrect (should be integer)", "error", err)
				}
				if propertyValue.Type.Is(openapi3.TypeInteger) {
					maxPtr := float64(max)
					propertyValue.Max = &maxPtr
				} else if propertyValue.Type.Is(openapi3.TypeString) {
					//nolint:gosec // disable G115
					maxPtr := uint64(max)
					propertyValue.MaxLength = &maxPtr
				}
			}
		}

		// Description
		description, ok := field.Tag.Lookup("description")
		if ok {
			propertyValue.Description = description
		}
		jsonTag, ok := field.Tag.Lookup("json")
		if ok {
			if strings.Contains(jsonTag, ",omitempty") {
				propertyValue.Nullable = true
			}
		}
		propertyCopy.Value = &propertyValue

		schemaRef.Value.Properties[jsonFieldName] = &propertyCopy
	}
}

// dive returns a schemaTag which includes the generated openapi3.SchemaRef and
// the name of the struct being passed in.
// If the type is a pointer, map, channel, function, or unsafe pointer,
// it will dive into the type and return the name of the type it points to.
// If the type is a slice or array type it will dive into the type as well as
// build and openapi3.Schema where Type is array and Ref is set to the proper
// components Schema
func dive(openapi *OpenAPI, t reflect.Type, tag SchemaTag, maxDepth int) SchemaTag {
	if maxDepth == 0 {
		return SchemaTag{
			Name: "default",
			SchemaRef: openapi3.SchemaRef{
				Ref: "#/components/schemas/default",
			},
		}
	}

	switch t.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		return dive(openapi, t.Elem(), tag, maxDepth-1)

	case reflect.Slice, reflect.Array:
		item := dive(openapi, t.Elem(), tag, maxDepth-1)
		tag.Name = item.Name
		tag.Value = openapi3.NewArraySchema()
		tag.Value.Items = &item.SchemaRef
		return tag

	default:
		tag.Name = t.Name()
		if t.Kind() == reflect.Struct && strings.HasPrefix(tag.Name, "DataOrTemplate") {
			return dive(openapi, t.Field(0).Type, tag, maxDepth-1)
		}
		tag.Ref = "#/components/schemas/" + tag.Name
		tag.Value = openapi.getOrCreateSchema(tag.Name, reflect.New(t).Interface())

		return tag
	}
}

func SchemaTagFromType(openapi *OpenAPI, v any) SchemaTag {
	if v == nil {
		// ensure we add unknown-interface to our schemas
		schema := openapi.getOrCreateSchema("unknown-interface", struct{}{})
		return SchemaTag{
			Name: "unknown-interface",
			SchemaRef: openapi3.SchemaRef{
				Ref:   "#/components/schemas/unknown-interface",
				Value: schema,
			},
		}
	}

	return dive(openapi, reflect.TypeOf(v), SchemaTag{}, 5)
}
