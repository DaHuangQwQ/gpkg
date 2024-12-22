package ginx

import (
	"encoding/json"
	"errors"
	"github.com/DaHuangQwQ/gpkg/ginx/openapi"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
)

type Server struct {
	*gin.Engine
	OpenAPI openapi.OpenAPI
}

func NewServer(opts ...gin.OptionFunc) *Server {
	return &Server{
		Engine: gin.Default(opts...),
	}
}

func (s *Server) Handle(method, path string, handler gin.HandlerFunc) {
	s.Engine.Handle(method, path, handler)
}

func (s *Server) Start(addr string) error {
	return s.Engine.Run(addr)
}

func (s *Server) marshalSpec() ([]byte, error) {
	return json.MarshalIndent(s.OpenAPI.Description(), "", "	")
}

func (s *Server) saveOpenAPIToFile(path string) error {
	jsonFolder := filepath.Dir(path)

	err := os.MkdirAll(jsonFolder, 0o750)
	if err != nil {
		return errors.New("error creating docs directory")
	}

	f, err := os.Create(path)
	if err != nil {
		return errors.New("error creating file")
	}
	defer f.Close()

	marshal, err := json.Marshal(s.OpenAPI.Description())
	if err != nil {
		return err
	}

	_, err = f.Write(marshal)
	if err != nil {
		return errors.New("error writing file ")
	}

	return nil
}

// Registers the routes to serve the OpenAPI spec and Swagger UI.
func (s *Server) registerOpenAPIRoutes(path string) {
	s.GET(path, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, s.OpenAPI.Description())
	})
}
