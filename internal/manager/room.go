package manager

import (
	"jungle/server/internal/game"
	"jungle/server/pkg/event"
	"jungle/server/pkg/transfer"
	"jungle/server/pkg/utils"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type gameRoom struct {
	game       *game.Game
	clientList []*ClientInfo

	nameList []string
}

func NewRoom() *gameRoom {
	r := new(gameRoom)

	r.clientList = make([]*ClientInfo, 0, PLAYER_PER_ROOM)
	r.nameList = make([]string, 0, PLAYER_PER_ROOM)

	return r
}

func (gr *gameRoom) addClient(conn *websocket.Conn, name string) (bool, string) {
	newClient := ClientInfo{
		conn:  conn,
		sid:   utils.RandomString(SID_LENGTH),
		name:  name,
		index: len(gr.clientList),
	}

	gr.clientList = append(gr.clientList, &newClient)
	gr.nameList = append(gr.nameList, newClient.name)

	gr.lobbyBroadcast(transfer.LOBBY_WAIT)

	if len(gr.clientList) == PLAYER_PER_ROOM {
		return true, newClient.sid
	}

	return false, newClient.sid
}

func (gr *gameRoom) removeClient(sid string) {
	gr.game.RemovePlayer(sid)
}

func (gr *gameRoom) startGame() {
	gr.lobbyBroadcast(transfer.LOBBY_START)

	time.Sleep(time.Second)

	sIdList := make([]string, PLAYER_PER_ROOM)
	for i, client := range gr.clientList {
		sIdList[i] = client.sid
	}
	gr.game = game.NewGame(sIdList)
	gr.game.EventManager.On(transfer.DISPATCH_MAP_INFO, func(msg event.Message) {
		d := msg.(transfer.ShareInfo)
		targets := d.Target
		transObj := transfer.TransferObj{
			Type:   transfer.SEND_GAMESHARE,
			Result: d,
		}

		logger.Info("ShareInfo", zap.String("data", d.String()))

		gr.gameBroadcast(targets, transObj)
	})
	gr.game.EventManager.On(transfer.DISPATCH_COMPANY_INFO, func(msg event.Message) {
		d := msg.(transfer.CompanyInfo)
		targets := d.Target
		transObj := transfer.TransferObj{
			Type:   transfer.SEND_GAMECOMPANY,
			Result: d,
		}

		logger.Info("CompanyInfo", zap.String("data", d.String()))

		gr.gameBroadcast(targets, transObj)
	})
	gr.game.EventManager.On(transfer.DISPATCH_ADMIT_ACTION, func(msg event.Message) {
		d := msg.(transfer.AdmitAction)
		targets := d.Target
		transObj := transfer.TransferObj{
			Type:   transfer.SEND_ADMITACTION,
			Result: d,
		}

		logger.Info("AdmitAction", zap.String("data", d.String()))

		gr.gameBroadcast(targets, transObj)
	})
	gr.game.EventManager.On(transfer.DISPATCH_COUNTDOWN, func(msg event.Message) {
		d := msg.(transfer.CountDown)
		targets := d.Target
		transObj := transfer.TransferObj{
			Type:   transfer.SEND_COUNTDOWN,
			Result: d,
		}

		gr.gameBroadcast(targets, transObj)
	})
	gr.game.EventManager.On(transfer.DISPATCH_END, func(msg event.Message) {
		d := msg.(transfer.EndScore)
		targets := d.Target
		transObj := transfer.TransferObj{
			Type:   transfer.SEND_GAMEEND,
			Result: d,
		}

		logger.Info("EndScore", zap.String("data", d.String()))

		gr.gameBroadcast(targets, transObj)
	})

	gr.handleAction()
}

func (gr *gameRoom) lobbyBroadcast(state string) {
	trans := transfer.TransferObj{
		Type: transfer.SEND_LOBBY,
	}

	for i := range gr.clientList {
		res := transfer.Lobby{
			Position: gr.nameList,
			Index:    i,
			State:    state,
		}

		trans.Result = res

		gr.clientList[i].send(trans)
	}
}

func (gr *gameRoom) gameBroadcast(targets []string, transObj interface{}) {
	for i := range gr.clientList {
		if utils.FindString(targets, gr.clientList[i].sid) != len(targets) {
			gr.clientList[i].send(transObj)
		}
	}
}

func (gr *gameRoom) handleAction() {
	for idx := range gr.clientList {
		go func(i int) {
			for {
				var msg transfer.ClientAction
				err := gr.clientList[i].conn.ReadJSON(&msg)
				if err != nil {
					logger.Warn(err.Error(), zap.String("timing", "Unmarshal ClientAction"))
					return
				}

				// From 要自己填
				msg.Result.From = gr.clientList[i].sid

				logger.Info("ClientAction", zap.String("RECEIVE", msg.String()))

				if msg.Type == transfer.TYPE_ACTION {
					gr.game.EventManager.Emit(transfer.RECEIVE_CLIENT_ACTION, msg)
				}

			}
		}(idx)
	}
}
