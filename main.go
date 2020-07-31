package main

import (
	"flag"
	"jungle/server/game/internal/manager"
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var addr = flag.String("addr", "localhost:8088", "http service address")
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var logger *zap.Logger

func init() {
	logger, _ = zap.NewDevelopment()
}

var gameManager *manager.GameManager

func main() {
	gameManager = manager.StartServer()

	http.HandleFunc("/", handleConnect)

	logger.Fatal(http.ListenAndServe(*addr, nil).Error())
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Warn(err.Error(), zap.String("timing", "Upgrade"))
		return
	}

	gameManager.HandleLogin(c)
}
