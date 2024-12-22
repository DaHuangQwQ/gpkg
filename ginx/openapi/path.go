package openapi

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"strings"
)

type Path[Res, Req any] struct {
	Operation            *openapi3.Operation
	FullName             string
	Path                 string
	AcceptedContentTypes []string
	DefaultStatusCode    int
	Method               string
	overrideDescription  bool
	Middlewares          []gin.HandlerFunc
}

func (p Path[Res, Req]) GenerateDefaultOperationID() {
	p.Operation.OperationID = p.Method + "_" + strings.ReplaceAll(strings.ReplaceAll(p.Path, "{", ":"), "}", "")
}

func (p Path[Res, Req]) GenerateDefaultDescription() {
	if p.overrideDescription {
		return
	}
	p.Operation.Description = DefaultDescription(p.FullName, p.Middlewares) + p.Operation.Description
}

func (p Path[Res, Req]) NameFromNamespace(human any) string {
	ss := strings.Split(p.FullName, ".")
	return ss[len(ss)-1]
}

func (p *Path[Res, Req]) RegisterOpenAPIOperation(openapi *OpenAPI) error {
	operation, err := registerOpenAPIOperation[Res, Req](openapi, *p)
	p.Operation = operation
	return err
}
