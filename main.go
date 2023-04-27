package main

import (
	"flag"
	"jungle/server/internal/manager"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var addr = flag.String("addr", "0.0.0.0:8088", "http service address")
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var logger *zap.Logger

func init() {
	logger, _ = zap.NewDevelopment()
}

var gameManager *manager.GameManager

func main() {
	gameManager = manager.NewGameManager()

	http.HandleFunc("/", handleConnect)

	logger.Fatal(http.ListenAndServe(*addr, nil).Error())
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Warn(err.Error(), zap.String("timing", "Upgrade"))
		return
	}

	// c.SetCloseHandler(func(code int, text string) error {
	// 	logger.Info(utils.MemUsageString(), zap.String("close", r.URL.Host))
	// 	return nil
	// })

	c.SetReadDeadline(time.Now().Add(time.Second * 3))

	gameManager.HandleLogin(c)
}
