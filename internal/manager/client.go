package manager

import (
	"sync"

	"github.com/gorilla/websocket"
)

const (
	SID_LENGTH = 6
)

type ClientInfo struct {
	conn  *websocket.Conn
	sid   string
	mutex sync.Mutex

	name  string
	index int
}

func (p *ClientInfo) send(data interface{}) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	return p.conn.WriteJSON(data)
}
