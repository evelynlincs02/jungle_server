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
	state      string
}

func NewRoom() *gameRoom {
	r := new(gameRoom)

	r.clientList = make([]*ClientInfo, 0, PLAYER_PER_ROOM)
	r.state = transfer.LOBBY_WAIT

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

	gr.lobbyBroadcast(transfer.LOBBY_WAIT)

	if len(gr.clientList) == PLAYER_PER_ROOM {
		return true, newClient.sid
	}

	return false, newClient.sid
}

func (gr *gameRoom) removeClient(sid string) {
	var cIdx int
	for i := range gr.clientList {
		if gr.clientList[i].sid == sid {
			cIdx = i
			break
		}
	}
	if gr.state == transfer.LOBBY_WAIT {
		nowNum := len(gr.clientList)
		gr.clientList[cIdx] = gr.clientList[nowNum-1]
		gr.clientList = gr.clientList[:nowNum-1]
		gr.lobbyBroadcast(transfer.LOBBY_WAIT)
	} else {
		gr.clientList[cIdx].name = "OFFLINE"
		gr.game.RemovePlayer(sid)
		gr.lobbyBroadcast(transfer.LOBBY_OFFLINE)
	}
	logger.Debug("Remove", zap.Int("idx", cIdx), zap.Strings("sids", gr.getClientList("sid")))
}

func (gr *gameRoom) startGame() {
	gr.state = transfer.LOBBY_START
	gr.lobbyBroadcast(transfer.LOBBY_START)

	time.Sleep(time.Second)

	sIdList := gr.getClientList("sid")
	gr.game = game.NewGame(sIdList)
	gr.game.EventManager.On(transfer.DISPATCH_MAP_INFO, func(msg event.Message) {
		d := msg.(transfer.ShareInfo)
		targets := d.Target
		transObj := transfer.TransferObj{
			Type:   transfer.SEND_GAMESHARE,
			Result: d,
		}

		logger.Debug(utils.MemUsageString(), zap.String(transfer.DISPATCH_MAP_INFO, d.String()))

		gr.gameBroadcast(targets, transObj)
	})
	gr.game.EventManager.On(transfer.DISPATCH_COMPANY_INFO, func(msg event.Message) {
		d := msg.(transfer.CompanyInfo)
		targets := d.Target
		transObj := transfer.TransferObj{
			Type:   transfer.SEND_GAMECOMPANY,
			Result: d,
		}

		logger.Debug(utils.MemUsageString(), zap.String(transfer.DISPATCH_COMPANY_INFO, d.String()))

		gr.gameBroadcast(targets, transObj)
	})
	gr.game.EventManager.On(transfer.DISPATCH_ADMIT_ACTION, func(msg event.Message) {
		d := msg.(transfer.AdmitAction)
		targets := d.Target
		transObj := transfer.TransferObj{
			Type:   transfer.SEND_ADMITACTION,
			Result: d,
		}

		logger.Debug(utils.MemUsageString(), zap.String(transfer.DISPATCH_ADMIT_ACTION, d.String()))

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

		logger.Debug(utils.MemUsageString(), zap.String(transfer.DISPATCH_END, d.String()))

		gr.gameBroadcast(targets, transObj)
	})

	gr.handleAction()
}

func (gr *gameRoom) lobbyBroadcast(state string) {
	trans := transfer.TransferObj{
		Type: transfer.SEND_LOBBY,
	}

	nameList := gr.getClientList("name")

	for i := range gr.clientList {
		res := transfer.Lobby{
			Position: nameList,
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

				logger.Debug(utils.MemUsageString(), zap.String("RECEIVE", msg.String()))

				if msg.Type == transfer.TYPE_ACTION {
					gr.game.EventManager.Emit(transfer.RECEIVE_CLIENT_ACTION, msg)
				}

			}
		}(idx)
	}
}

func (gr *gameRoom) getClientList(t string) []string {
	res := make([]string, 0, PLAYER_PER_ROOM)
	switch t {
	case "name":
		for _, c := range gr.clientList {
			res = append(res, c.name)
		}
		return res
	default:
		for _, c := range gr.clientList {
			res = append(res, c.sid)
		}
		return res
	}
}
