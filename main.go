package main

import (
	"encoding/json"
	"flag"
	"jungle/server/game/internal/game"
	"jungle/server/game/pkg/event"
	"jungle/server/game/pkg/transfer"
	"jungle/server/game/pkg/utils"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/websocket"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{
		ForceColors: true,
	})
	log.SetLevel(log.ErrorLevel)
}

var addr = flag.String("addr", "localhost:8088", "http service address")
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type clientInfo struct {
	conn *websocket.Conn
	sid  string
	mu   sync.Mutex
}

var clientList []clientInfo
var nameList []string
var endSignal chan struct{}

func main() {
	endSignal = make(chan struct{})

	defer func() {
		log.Trace("main END")
	}()

	clientList = make([]clientInfo, 0, 6)
	nameList = make([]string, 0, 6)

	http.HandleFunc("/", handleConnect)

	log.Fatal(http.ListenAndServe(*addr, nil))

	<-endSignal
}

func handleConnect(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warn("upgrade:", err)
		return
	}

	var msg transfer.Login
	err = c.ReadJSON(&msg)
	if err != nil {
		log.Warn("Unmarshal:", err)
		return
	}

	log.WithFields(log.Fields{"login": msg.Result.GiveName}).Infof("%s", msg.String())

	if msg.Type == transfer.TYPE_LOGIN {
		loginData := msg.Result
		pIdx := len(clientList)
		clientList = append(clientList, clientInfo{conn: c, sid: strconv.Itoa(pIdx)})

		name := loginData.GiveName
		nameList = append(nameList, name)

		loginObj := transfer.TransferObj{
			Type: transfer.SEND_LOBBY,
			Result: transfer.Lobby{
				Position: nameList,
				Index:    len(clientList) - 1,
				State:    transfer.LOBBY_WAIT,
			},
		}
		jsonByte, err := json.Marshal(loginObj)
		if err != nil {
			log.Warn("json.Marshal ERROR", err)
			return
		}
		c.WriteMessage(websocket.TextMessage, jsonByte)

		if len(clientList) == 6 {
			go startGame(clientList)
		}
	}
}

func startGame(cList []clientInfo) {
	idList := make([]string, 6)
	for i := range cList {
		idList[i] = cList[i].sid
		loginObj := transfer.TransferObj{
			Type: transfer.SEND_LOBBY,
			Result: transfer.Lobby{
				Position: nameList,
				Index:    i,
				State:    transfer.LOBBY_START,
			},
		}
		cList[i].send(websocket.TextMessage, loginObj)
	}

	time.Sleep(time.Second)
	jungleGame := game.NewGame(idList)
	jungleGame.EventManager.On(transfer.DISPATCH_MAP_INFO, func(msg event.Message) {
		sendGameData(cList, transfer.SEND_GAMESHARE, msg)
	})
	jungleGame.EventManager.On(transfer.DISPATCH_COMPANY_INFO, func(msg event.Message) {
		sendGameData(cList, transfer.SEND_GAMECOMPANY, msg)
	})
	jungleGame.EventManager.On(transfer.DISPATCH_ADMIT_ACTION, func(msg event.Message) {
		sendGameData(cList, transfer.SEND_ADMITACTION, msg)
	})
	jungleGame.EventManager.On(transfer.DISPATCH_COUNTDOWN, func(msg event.Message) {
		sendGameData(cList, transfer.SEND_COUNTDOWN, msg)
	})
	jungleGame.EventManager.On(transfer.DISPATCH_END, func(msg event.Message) {
		sendGameData(cList, transfer.SEND_GAMEEND, msg)
		endSignal <- struct{}{}
	})

	for idx := range cList {
		go func(i int) {
			for {
				var msg transfer.ClientAction
				err := cList[i].conn.ReadJSON(&msg)
				if err != nil {
					log.Warn("Unmarshal:", err)
					return
				}

				// From 要自己填
				msg.Result.From = cList[i].sid
				log.WithFields(log.Fields{"RECEIVE": "ClientAction"}).Infoln(msg.String())

				if msg.Type == transfer.TYPE_ACTION {
					jungleGame.EventManager.Emit(transfer.RECEIVE_CLIENT_ACTION, msg)
				}

			}
		}(idx)
	}
}

func sendGameData(cList []clientInfo, dataType string, data event.Message) {
	transObj := transfer.TransferObj{
		Type:   dataType,
		Result: data,
	}

	var targets []string

	switch dataType {
	case transfer.SEND_GAMESHARE:
		d := data.(transfer.ShareInfo)
		targets = d.Target
		log.WithFields(log.Fields{dataType: "ShareInfo"}).Infoln(d.String())
	case transfer.SEND_GAMECOMPANY:
		d := data.(transfer.CompanyInfo)
		targets = d.Target
		log.WithFields(log.Fields{dataType: "CompanyInfo"}).Infoln(d.String())
	case transfer.SEND_COUNTDOWN:
		d := data.(transfer.CountDown)
		targets = d.Target
	case transfer.SEND_ADMITACTION:
		d := data.(transfer.AdmitAction)
		targets = d.Target
		log.WithFields(log.Fields{dataType: "AdmitAction"}).Infoln(d.String())
	case transfer.SEND_GAMEEND:
		d := data.(transfer.EndScore)
		targets = d.Target
		log.WithFields(log.Fields{dataType: "EndScore"}).Infoln(d.String())
	}

	for i := range cList {
		if utils.FindString(targets, cList[i].sid) != len(targets) {
			cList[i].send(websocket.TextMessage, transObj)
		}
	}
}

func (p *clientInfo) send(messageType int, data interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.conn.WriteJSON(data)
}
