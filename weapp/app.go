package weapp

import (
	"gpkg/ginx"
	"gpkg/grpcx"
	"gpkg/saramax"
)

type App struct {
	GRPCServer *grpcx.Server
	WebServer  *ginx.Server
	Consumers  []saramax.Consumer
}
