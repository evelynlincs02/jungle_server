package manager

import (
	"jungle/server/game/pkg/transfer"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	var err error
	logger, err = config.Build()
	if err != nil {
		panic(err)
	}
}

const (
	PLAYER_PER_ROOM = 6
)

type GameManager struct {
	gameRoom map[int]*gameRoom
	// clientMap map[string]*websocket.Conn

	roomIndex int
}

func StartServer() *GameManager {
	gm := new(GameManager)

	gm.gameRoom = make(map[int]*gameRoom, 10)
	gm.roomIndex = -1
	gm.nextRoom()

	logger.Info("Listening")

	return gm
}

func (gm *GameManager) nextRoom() {
	gm.roomIndex++
	gm.gameRoom[gm.roomIndex] = NewRoom()
}

func (gm *GameManager) HandleLogin(c *websocket.Conn) {
	var msg transfer.Login
	err := c.ReadJSON(&msg)
	if err != nil {
		logger.Warn(err.Error(), zap.String("timing", "Unmarshal Login"))
		return
	}

	logger.Info(msg.Result.GiveName, zap.String("timing", "Login"))

	waitingRoom := gm.gameRoom[gm.roomIndex]
	loginData := msg.Result
	full := waitingRoom.addClient(c, loginData.GiveName)
	if full {
		go waitingRoom.startGame()
		gm.nextRoom()
	}
}
