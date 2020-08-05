package manager

import (
	"jungle/server/pkg/transfer"
	"jungle/server/pkg/utils"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	config := zap.NewDevelopmentConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	config.DisableCaller = true

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

	roomIndex int
}

func NewGameManager() *GameManager {
	gm := new(GameManager)

	gm.gameRoom = make(map[int]*gameRoom, 10)
	gm.roomIndex = -1
	gm.nextRoom()

	logger.Debug("Listening")

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

	waitingRoom := gm.gameRoom[gm.roomIndex]
	loginData := msg.Result
	full, sid := waitingRoom.addClient(c, loginData.GiveName)
	if full {
		go waitingRoom.startGame()
		gm.nextRoom()
	}

	logger.Info(utils.MemUsageString(),
		zap.String("login name", msg.Result.GiveName), zap.String("login sid", sid), zap.Strings("players", waitingRoom.getClientList("sid")))

	c.SetCloseHandler(func(code int, text string) error {
		logger.Info(utils.MemUsageString(), zap.String("close", sid))
		waitingRoom.removeClient(sid)
		return nil
	})
}
