package logsv

import (
	"encoding/json"
	"log/slog"
	"os"
	"slices"

	"github.com/dwdwow/props"
	"github.com/dwdwow/ws/wssv"
)

type Log struct {
	Level  string `json:"level"`
	ApiKey string `json:"apiKey"`
}

type ClientMsgEvent string

const (
	ClientMsgNewLog ClientMsgEvent = "new_log"
	ClientMsgSubLog ClientMsgEvent = "sub_log"
)

type ClientMsg struct {
	Event  ClientMsgEvent `json:"event"`
	Level  string         `json:"level"`
	ApiKey string         `json:"apiKey"`
	LogStr string         `json:"logStr"`
}

type Server struct {
	sv *wssv.Server

	watchers *props.SafeRWMap[string, []*wssv.Client]

	logger *slog.Logger
}

func NewServer(port uint64, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	sv := wssv.NewServer(port, logger)

	watchers := props.NewSafeRWMap[string, []*wssv.Client]()

	return &Server{
		sv:       sv,
		watchers: watchers,
		logger:   logger,
	}
}

func (sv *Server) Start() {
	err := sv.sv.Start(func(server *wssv.Server, client *wssv.Client) {
		sv.logger.Info("New Connection")
	}, func(server *wssv.Server, client *wssv.Client, messageType int, msg []byte, err error) {
		if err != nil {
			sv.removeClient(client)
			return
		}
		if len(msg) == 0 {
			sv.logger.Error("Message Is Empty")
			return
		}
		cltMsg := new(ClientMsg)
		err = json.Unmarshal(msg, cltMsg)
		if err != nil {
			sv.logger.Error("Cannot Unmarshal Client Msg", "msg", string(msg), "err", err)
			return
		}
		key := cltMsg.ApiKey + cltMsg.Level
		switch cltMsg.Event {
		case ClientMsgSubLog:
			sv.watchers.Lock()
			sv.watchers.Data[key] = append(sv.watchers.Data[key], client)
			sv.watchers.Unlock()
		case ClientMsgNewLog:
			sv.watchers.Lock()
			clts := sv.watchers.Data[key]
			for _, clt := range clts {
				clt := clt
				go func() {
					err := clt.WriteText([]byte(cltMsg.LogStr))
					if err != nil {
						sv.logger.Error("Cannot Write Text", "err", err)
					}
				}()
			}
			sv.watchers.Unlock()
		}
	})
	if err != nil {
		sv.logger.Error("Can Not Start", "err", err)
	}
}

func (sv *Server) removeClient(client *wssv.Client) {
	sv.watchers.Lock()
	defer sv.watchers.Unlock()
	sv.logger.Info("Removing Connection")
	for k, clts := range sv.watchers.Data {
		for i, clt := range clts {
			if clt == client {
				_ = clt.Conn.Close()
				newClts := slices.Delete(clts, i, i)
				sv.watchers.Data[k] = newClts
				return
			}
		}
	}
}
