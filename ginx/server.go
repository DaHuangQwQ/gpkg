package ginx

import "github.com/gin-gonic/gin"

type Server struct {
	*gin.Engine
	Addr string
}

func NewServer(addr string, opts ...gin.OptionFunc) *Server {
	return &Server{
		Engine: gin.Default(opts...),
		Addr:   addr,
	}
}

func (s *Server) Handle(method, path string, handler gin.HandlerFunc) {
	s.Engine.Handle(method, path, handler)
}

func (s *Server) Start() error {
	return s.Engine.Run(s.Addr)
}
