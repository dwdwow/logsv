package logsv

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/dwdwow/props"
	"github.com/dwdwow/ws/wsclt"
	"github.com/gorilla/websocket"
)

func TestServer_Start(t *testing.T) {
	apiKey := "hsladsdgh143kladg4859"

	server := NewServer(8899, nil)

	go func() {
		server.Start()
	}()

	time.Sleep(time.Second * 3)

	go func() {
		logClt := wsclt.NewBaseClient("ws://127.0.0.1:8899", true, 100, nil)
		err := logClt.Start()
		props.PanicIfNotNil(err)
		for {
			time.Sleep(time.Second)

			err := logClt.WriteJSON(ClientMsg{
				Event:  ClientMsgNewLog,
				Level:  "ERROR",
				ApiKey: apiKey,
				LogStr: strconv.FormatInt(time.Now().UnixMilli(), 10),
			})
			props.PanicIfNotNil(err)
		}
	}()

	go func() {
		clt := wsclt.NewBaseClient("ws://127.0.0.1:8899", true, 100, nil)
		err := clt.Start()
		props.PanicIfNotNil(err)
		err = clt.WriteJSON(ClientMsg{
			Event:  ClientMsgSubLog,
			Level:  "ERROR",
			ApiKey: apiKey,
			LogStr: "",
		})
		props.PanicIfNotNil(err)
		for {
			msgType, msgData, err := clt.Read()
			props.PanicIfNotNil(err)
			if msgType != websocket.TextMessage {
				panic("not a text message")
			}
			fmt.Println(string(msgData))
		}
	}()

	select {}
}
