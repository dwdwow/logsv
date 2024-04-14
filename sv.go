package logsv

import (
	"log/slog"
	"os"

	"github.com/dwdwow/props"
	"github.com/dwdwow/ws/wssv"
)

type Log struct {
	Level  int
	ApiKey string
}

type Server struct {
	sv *wssv.Server

	loggers, clts props.SafeRWSlice[*wssv.Client]

	logger *slog.Logger
}

func NewServer(port uint64, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	sv := wssv.NewServer(port, logger)

	return &Server{
		sv:     sv,
		logger: logger,
	}
}

func (sv *Server) Start() {
	err := sv.sv.Start(func(server *wssv.Server, client *wssv.Client) {
		sv.clts.Append(client)

	}, func(server *wssv.Server, client *wssv.Client, messageType int, msg []byte, err error) {

	})
	if err != nil {
		sv.logger.Error("Can Not Start", "err", err)
	}
}
