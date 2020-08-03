package transfer

import (
	"fmt"
	"strconv"
)

// ----------------Transfer to client
type TransferObj struct {
	Type   string      `json:"type"`
	Result interface{} `json:"result"`
}

type Lobby struct {
	Position []string `json:"position"`
	State    string   `json:"state"`
	Index    int      `json:"index"`
}

type ShareInfo struct {
	Target         []string  `json:"target"`
	Month          int       `json:"month"`
	Market         [16]int   `json:"market"`
	BearOffice     []bool    `json:"bear_office"`
	BearProgress   []string  `json:"bear_progress"`
	DeerOffice     []bool    `json:"deer_office"`
	DeerProgress   []string  `json:"deer_progress"`
	PlayerPosition [6]string `json:"player_position"`
	ActionPoint    int       `json:"action_point"`

	DrawCard *DrawCard `json:"draw_card,omitempty"`
}
type DrawCard struct {
	CardType string `json:"card_type"`
	Card     []int  `json:"card"`
}

func (info *ShareInfo) String() string {
	ret := fmt.Sprintf("Market=%v, Position=%v, BearOffice= %v, BearProgress=%v, DeerOffice=%v, DeerProgress=%v",
		info.Market, info.PlayerPosition, info.BearOffice, info.BearProgress, info.DeerOffice, info.DeerProgress)
	if info.DrawCard != nil {
		ret += fmt.Sprintf(", DrawCard=%v %v", info.DrawCard.CardType, info.DrawCard.Card)
	}
	return ret
}

type CompanyInfo struct {
	Target      []string `json:"target"`
	Company     string   `json:"company"`
	ProductId   []int    `json:"product_id"`
	HandCard    [3][]int `json:"design_card"`
	CheckMarket []string `json:"check_market"`
	EndOfMonth  *int     `json:"end_of_month"`
}

func (info *CompanyInfo) String() string {
	return fmt.Sprintf("Company=%v, ProductId=%v, HandCard= %v, CheckMarket= %v",
		info.Company, info.ProductId, info.HandCard, info.CheckMarket)
}

type AdmitAction struct {
	Target []string `json:"target"`
	Action []string `json:"action"`

	DropType   *string  `json:"drop_type,omitempty"`
	DropMarket *[][]int `json:"drop_market,omitempty"`
}

func (info *AdmitAction) String() string {
	ret := fmt.Sprintf("Target=%v, Action=%v",
		info.Target, info.Action)
	if info.DropType != nil {
		ret += fmt.Sprintf(", DropType=%v, DropMarket=%v",
			*info.DropType, info.DropMarket)
	}
	return ret
}

type EndScore struct {
	Target    []string     `json:"target"`
	BearScore []ProductSum `json:"bear_score"`
	DeerScore []ProductSum `json:"deer_score"`
}
type ProductSum struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
	Num   int    `json:"num"`
}

func (info *ProductSum) String() string {
	return fmt.Sprintf("%s:%d", info.Name+strconv.Itoa(info.Value), info.Num)
}

func (info *EndScore) String() string {
	bs := ""
	for _, s := range info.BearScore {
		bs += fmt.Sprintf("{%s}", s.String())
	}
	ds := ""
	for _, s := range info.DeerScore {
		ds += fmt.Sprintf("{%s}", s.String())
	}
	return fmt.Sprintf("BearScore={%s}, DeerScore={%s}", bs, ds)
}

type CountDown struct {
	Target      []string `json:"target"`
	Count       int      `json:"count"`
	PlayerIndex int      `json:"player_index"`
	CountType   string   `json:"count_type"`
}

// ----------------

// ----------------Receive from client
type ClientAction struct {
	Type   string `json:"type"`
	Result struct {
		From       string    `json:"from"`
		ActionType string    `json:"action_type"`
		Data       *[]string `json:"data,omitempty"`
	}
}

func (info *ClientAction) String() string {
	return fmt.Sprintf("From=%s, ActionType=%s, Data=%v", info.Result.From, info.Result.ActionType, info.Result.Data)
}

type Login struct {
	Type   string `json:"type"`
	Result struct {
		// Id       string `json:"id"`
		GiveName string `json:"give_name"`
	}
}

func (info *Login) String() string {
	return fmt.Sprintf("Type=%s, GiveName=%s", info.Type, info.Result.GiveName)
}

// ----------------
const (
	DISPATCH_MAP_INFO     = "MAP_INFO"
	DISPATCH_COMPANY_INFO = "COMPANY_INFO"
	DISPATCH_ADMIT_ACTION = "ADMIT_ACTION"
	DISPATCH_COUNTDOWN    = "COUNTDOWN"
	DISPATCH_END          = "END"

	RECEIVE_CLIENT_ACTION = "CLIENT_ACTION"

	LOBBY_START = "starting"
	LOBBY_WAIT  = "waiting"

	TYPE_LOGIN       = "login"
	TYPE_ACTION      = "action"
	SEND_LOBBY       = "lobby"
	SEND_GAMESHARE   = "game_share"
	SEND_GAMECOMPANY = "game_company"
	SEND_ADMITACTION = "game_player"
	SEND_COUNTDOWN   = "countdown"
	SEND_GAMEEND     = "game_end"
)
