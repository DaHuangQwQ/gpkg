package ginx

import "github.com/gin-gonic/gin"

type Server struct {
	*gin.Engine
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
