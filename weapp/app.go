package weapp

import (
	"github.com/DaHuangQwQ/gpkg/ginx"
	"github.com/DaHuangQwQ/gpkg/grpcx"
	"github.com/DaHuangQwQ/gpkg/saramax"
)

type App struct {
	GRPCServer *grpcx.Server
	WebServer  *ginx.Server
	Consumers  []saramax.Consumer
}
