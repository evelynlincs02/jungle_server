package game

import (
	"jungle/server/game/internal/cardpool"
	"jungle/server/game/internal/company"
	"jungle/server/game/internal/junglemap"
	"jungle/server/game/pkg/event"
	"jungle/server/game/pkg/transfer"
	"jungle/server/game/pkg/utils"
	"sort"
	"strconv"
	"strings"
	"time"
)

type autostr uint

const (
	ROUND_SECOND = 60
	DROP_SECOND  = 10
	MAX_PLAYER   = 6
	LAST_MONTH   = 9
	MAX_HAND     = 4

	// 方便直接靠編輯器帶入，減少打錯字的機會
	BEAR        = "bear"
	DEER        = "deer"
	PUBLISH     = "publish"
	MARKET      = "market"
	HAND        = "hand"
	IDEA        = "idea"
	HAPPY       = "happy"
	CUBE        = "cube"
	OVER        = "over"
	MOVE        = "move"
	INTERACTION = "interaction"
	VISUAL      = "visual"
	CLARIFIED   = "clarified"
	CHECK       = "check"
	GIVE        = "give"
	DROP        = "drop"
)

type Game struct {
	cNames [2]string

	players    []string
	pOnline    []bool
	pNow       int
	pActPoint  int
	pSkillUsed bool

	jungleMap *junglemap.JungleMap
	company   map[string]*company.Company
	devMap    map[string]*company.DevMap

	changePool   *cardpool.CardPool
	obstaclePool *cardpool.CardPool

	EventManager event.EventEmitter
	ticker       *time.Ticker
	dropTicker   []*time.Ticker
}

func NewGame(p []string) *Game {
	g := new(Game)

	g.cNames = [2]string{BEAR, DEER}

	g.players = make([]string, MAX_PLAYER)
	g.pOnline = make([]bool, MAX_PLAYER)
	for i := range g.pOnline {
		g.players[i] = p[i]
		g.pOnline[i] = true
	}
	g.pNow = -1
	g.pActPoint = 0
	g.pSkillUsed = false

	g.jungleMap = junglemap.NewJungleMap()
	g.company = make(map[string]*company.Company)
	g.devMap = make(map[string]*company.DevMap)
	for i, n := range g.cNames {
		g.company[n] = company.NewCompany([]string{p[i], p[i+2], p[i+4]})
		g.devMap[n] = company.NewDevMap()
	}

	g.changePool = cardpool.NewChangePool()
	g.obstaclePool = cardpool.NewObstaclePool()

	g.EventManager = make(event.EventEmitter, 0)
	g.EventManager.On(transfer.RECEIVE_CLIENT_ACTION, func(msg event.Message) {
		action := msg.(transfer.ClientAction).Result
		if action.ActionType == DROP {
			g.parseDrop(action.From, *action.Data)
		} else {
			g.parseAction(action.ActionType, *action.Data...)
		}
	})

	g.dropTicker = make([]*time.Ticker, 6)

	go func() {
		time.Sleep(100 * time.Millisecond)
		g.start()
	}()

	return g
}

func (g *Game) start() {
	g.jungleMap.Init()
	for _, n := range g.cNames {
		g.company[n].Init()
		g.devMap[n].Init()

		g.EventManager.Emit(transfer.DISPATCH_COMPANY_INFO, g.makeCompanyInfo(n))
	}

	g.EventManager.Emit(transfer.DISPATCH_MAP_INFO, g.makeMapInfo())

	time.Sleep(3 * time.Second)

	// 抽變動卡
	for i := 0; i < 4; i++ {
		g.drawChange(true)
		// 在遊戲準備階段時，若任一辦公室棋盤格有連續變動，玩家均無需抽取阻礙卡
	}
	g.switchPlayer()
}

func (g *Game) removePlayer(p string) {
	pIdx := utils.FindString(g.players, p)
	if pIdx == MAX_PLAYER {
		return
	}
	pOfC := pIdx / 2
	cName := g.cNames[pIdx%2]
	comp := g.company[cName]

	// g.players[pIdx] = ""
	g.pOnline[pIdx] = false
	g.jungleMap.GoPos(pIdx, "Unknown")

	for i := len(comp.AllHand()[pOfC]); i > 0; i-- {
		comp.DropHand(pOfC, 0)
	}

	g.EventManager.Emit(transfer.DISPATCH_COMPANY_INFO, g.makeCompanyInfo(cName))

	if pIdx == g.pNow && g.ticker != nil {
		g.ticker.Stop()

		g.pActPoint = 0

		g.drawShare()
	}

}

func (g *Game) drawShare() {
	obstacle := g.drawChange(false)
	if obstacle != "" {
		over := g.drawObstacle(obstacle)
		if over {
			g.switchPlayer()
		}
	} else {
		g.switchPlayer()
	}
}

func (g *Game) drawChange(bFirst bool) string {
	var cId []int
	var cNow string
	if bFirst { // 第一次市場變動時不要抽"抽阻礙"的卡
		cId = g.changePool.Draw(1, 16)
		cNow = ""
	} else {
		cId = g.changePool.Draw(1)
		cNow = g.cNames[g.pNow%2]
	}
	detail := g.changePool.GetContent(cId[0])
	g.changePool.DropById(cId[0])

	obstacle := g.jungleMap.Change(detail, cNow)

	g.EventManager.Emit(transfer.DISPATCH_MAP_INFO, g.makeMapInfo(transfer.DrawCard{CardType: "change", Card: cId}))

	time.Sleep(3 * time.Second)

	return obstacle
}

func (g *Game) drawObstacle(cTarget string) bool {
	cId := g.obstaclePool.Draw(1)
	detail := g.obstaclePool.GetContent(cId[0])
	g.obstaclePool.DropById(cId[0])

	var pTarget int
	if g.cNames[g.pNow%2] == cTarget { // 需要抽阻礙的不是pNow這隊的時候，是找下一位對方玩家
		pTarget = g.pNow
	} else {
		pTarget = (g.pNow + 1) % 6
	}

	roundOver := true
	for k, v := range detail {
		kk := strings.Split(k, "_")
		if kk[0] == "p" {
			schedules := g.devMap[cTarget].Schedules()
			candidate := make([]int, 0, len(schedules))
			for i, sche := range schedules {
				if i == 0 { // 丟棄專案進度不包含license
					continue
				}
				if kk[1] == strings.Split(sche, "_")[0] {
					candidate = append(candidate, i)
				}
			}

			if len(candidate) > 1 {
				roundOver = false
				g.askDrop(pTarget, kk[1])
			} else if len(candidate) == 1 {
				g.devMap[cTarget].DropSchedule(g.company[cTarget], candidate[0])
			}

		} else if kk[0] == "m" {
			drop := g.jungleMap.DropOption([]string{kk[1]})
			drop[0] = append(drop[0], drop[1]...)
			if len(drop[0]) > v {
				g.jungleMap.DropMarket(drop[0][:v])
			} else {
				g.jungleMap.DropMarket(drop[0])
			}

		} else if kk[0] == "h" {
			comp := g.company[cTarget]
			pOfC := pTarget / 2
			for ; v > 0; v-- {
				comp.DropHand(pOfC, utils.FindString(comp.HandDetail(pOfC), kk[1]))
			}

			g.EventManager.Emit(transfer.DISPATCH_COMPANY_INFO, g.makeCompanyInfo(cTarget))

		} else { // k == "office"
			g.jungleMap.MakeHappy(cTarget, false, 3, 4, 5)
		}
	}

	g.EventManager.Emit(transfer.DISPATCH_MAP_INFO, g.makeMapInfo(transfer.DrawCard{
		CardType: "obstacle",
		Card:     cId,
	}))

	time.Sleep(3 * time.Second)

	return roundOver
}

func (g *Game) switchPlayer() {
	g.pNow++
	if g.pNow >= 6 {
		g.pNow = 0
	}
	g.pActPoint = 3
	g.pSkillUsed = false

	g.resetTimer()
}

func (g *Game) resetTimer() {
	if !g.pOnline[g.pNow] { // 玩家已離線
		g.pActPoint = 0
		g.drawShare()

		return
	}

	g.ticker = time.NewTicker(time.Second) // 相當於每秒一次送時間進 channel

	go func() { // receiver
		count := 0
		for {
			_ = <-g.ticker.C
			if count++; count > ROUND_SECOND {
				g.ticker.Stop() // 讓ticker不要再送
				g.parseAction(OVER)
				return
			}
			g.EventManager.Emit(transfer.DISPATCH_COUNTDOWN, transfer.CountDown{
				Target:      g.players,
				Count:       ROUND_SECOND - count,
				PlayerIndex: g.pNow,
				CountType:   "action",
			})
		}
	}()

	time.Sleep(time.Second)

	g.EventManager.Emit(transfer.DISPATCH_MAP_INFO, g.makeMapInfo())
	g.admitAction()
}

func (g *Game) admitAction() {
	cNow := g.cNames[g.pNow%2]
	pOfC := g.pNow / 2
	dm := g.devMap[cNow]

	canDo := make([]string, 0)

	canDo = append(canDo, OVER, MOVE)
	if dm.NumOfNextSchedule(CUBE) > 0 && g.jungleMap.NumOfHappy(cNow, 3, 4, 5) > 0 {
		canDo = append(canDo, CUBE)
	}
	if g.jungleMap.PosCanHappy(cNow, g.pNow) || (pOfC == 0 && !g.pSkillUsed && g.jungleMap.NumOfHappy(cNow) < 6) {
		canDo = append(canDo, HAPPY)
	}
	if g.jungleMap.PosCanMarket(cNow, g.pNow) {
		canDo = append(canDo, CHECK)
	}
	pHand := g.company[cNow].HandDetail(pOfC)
	if g.jungleMap.CheckSamePos(g.pNow) && len(pHand) > 0 {
		canDo = append(canDo, GIVE)
	}
	for i, h := range pHand {
		if h == PUBLISH && dm.NumOfNextSchedule(h) > 1 && dm.CheckLisense() && g.pActPoint == 3 && g.jungleMap.NumOfHappy(cNow, 0) == 1 {
			canDo = append(canDo, h)
		} else if (h == IDEA) || (h == HAPPY && g.jungleMap.NumOfHappy(cNow) < 6) {
			canDo = append(canDo, HAND+"_"+strconv.Itoa(i))
		} else if dm.NumOfNextSchedule(h) > 0 {
			if h == CLARIFIED && g.jungleMap.NumOfHappy(cNow, 1) == 1 {
				canDo = append(canDo, HAND+"_"+strconv.Itoa(i))
			} else if (h == INTERACTION || h == VISUAL) && g.jungleMap.NumOfHappy(cNow, 2) == 1 {
				canDo = append(canDo, HAND+"_"+strconv.Itoa(i))
			}
		}
	}

	canDo = utils.Unique(canDo)

	g.EventManager.Emit(transfer.DISPATCH_ADMIT_ACTION, transfer.AdmitAction{
		Target: []string{g.players[g.pNow]},
		Action: canDo,
	})
}

func (g *Game) parseAction(act string, args ...string) {
	cNname := g.cNames[g.pNow%2]
	pOfC := g.pNow / 2
	dm := g.devMap[cNname]
	comp := g.company[cNname]

	if act == PUBLISH {
		g.pActPoint -= 3
		comp.DropHand(pOfC, utils.FindString(comp.HandDetail(pOfC), PUBLISH))

		intargs := make([]int, len(args))
		for i, v := range args {
			intargs[i], _ = strconv.Atoi(v)
		}

		total, pType := dm.Publish(intargs, g.jungleMap.MarketSum())
		comp.Publish(len(intargs))

		g.EventManager.Emit(transfer.DISPATCH_COMPANY_INFO, g.makeCompanyInfo(cNname))

		if total >= 4 {
			g.endGame()
		}

		opt := g.jungleMap.DropOption(pType)
		if len(opt[0]) > 0 || len(opt[1]) > 0 {
			g.askDrop(g.pNow, MARKET, opt...) // 詢問棄市場
			g.ticker.Stop()
			return // 先返回，在接到玩家回應後再往下。若玩家時間到了仍未選完，客端會送已選的部分來 -> parseDrop
		}

	} else if strings.Split(act, "_")[0] == HAND {
		g.pActPoint -= 1
		hIdx, _ := strconv.Atoi(strings.Split(act, "_")[1])
		hName := comp.HandDetail(pOfC)[hIdx]

		cTarget, _ := strconv.Atoi(args[0])

		if hName == HAPPY {
			g.jungleMap.MakeHappy(cNname, true, cTarget)
			comp.DropHand(pOfC, hIdx)
		} else if hName == IDEA {
			dm.NewIdea(comp, cTarget)
			comp.DropHand(pOfC, hIdx)

		} else { // 專案進度卡
			dm.GoSchedule(cTarget) // 相信客端
			if pOfC == 1 && !g.pSkillUsed && (hName == INTERACTION || hName == VISUAL) {
				g.pSkillUsed = true
				g.pActPoint += 1
			}

			comp.GiveHand(pOfC, -1, hIdx) // 卡先放到某個地方，還沒要進棄排堆
		}

		g.EventManager.Emit(transfer.DISPATCH_COMPANY_INFO, g.makeCompanyInfo(cNname))

	} else if act == CUBE {
		g.pActPoint -= 1
		for _, v := range args {
			arg := strings.Split(v, "_")
			pIdx, _ := strconv.Atoi(arg[0])
			pCube, _ := strconv.Atoi(arg[1])

			dm.CubeUpdate(pIdx, pCube)
		}
		g.EventManager.Emit(transfer.DISPATCH_COMPANY_INFO, g.makeCompanyInfo(cNname))

	} else if act == MOVE {
		g.pActPoint -= 1
		g.jungleMap.GoPos(g.pNow, args[0])

	} else if act == CHECK {
		if pOfC == 2 && !g.pSkillUsed {
			g.pSkillUsed = true
		} else {
			g.pActPoint -= 1
		}
		mIdx, _ := strconv.Atoi(args[0])
		g.jungleMap.CheckMarket(cNname, mIdx)
		g.EventManager.Emit(transfer.DISPATCH_COMPANY_INFO, g.makeCompanyInfo(cNname))

	} else if act == HAPPY {
		g.pActPoint -= 1
		oIdx, _ := strconv.Atoi(args[0])
		if pOfC == 0 && !g.pSkillUsed && !g.jungleMap.PosCanHappy(cNname, g.pNow) {
			g.pSkillUsed = true
		}
		g.jungleMap.MakeHappy(cNname, true, oIdx)

	} else if act == GIVE {
		g.pActPoint -= 1
		arg := strings.Split(args[0], "_")
		hIdx, _ := strconv.Atoi(arg[0]) // hand index
		pIdx, _ := strconv.Atoi(arg[1]) // give to (team idx)

		needDrop := comp.GiveHand(pOfC, pIdx, hIdx)
		g.EventManager.Emit(transfer.DISPATCH_COMPANY_INFO, g.makeCompanyInfo(cNname))

		if needDrop {
			g.askDrop(pIdx*2+g.pNow%2, HAND)
		}

	} else if act == OVER {
		g.pActPoint = 0
	}

	g.EventManager.Emit(transfer.DISPATCH_MAP_INFO, g.makeMapInfo())

	if g.pActPoint == 0 {
		g.ticker.Stop()

		month, needDrop := comp.DrawHand(pOfC)

		time.Sleep(100 * time.Millisecond) // 用了手牌送做完動作的結果時機跟抽新牌的時機太近了會撞在一起，拉個間隔

		g.EventManager.Emit(transfer.DISPATCH_COMPANY_INFO, g.makeCompanyInfo(cNname, month))

		time.Sleep(4 * time.Second) // wait client show anime of hand card

		g.jungleMap.AddMonth(month)
		if g.jungleMap.Month() > LAST_MONTH {
			g.endGame()
			return
		}

		if needDrop {
			g.askDrop(g.pNow, HAND)
		} else {
			g.drawShare()
		}

	} else {
		g.admitAction()
	}
}

/*
給玩家選棄牌
 HAND
	1) 玩家回合結束抽牌時超過 -> 玩家選
	2) 玩家給隊友牌使得隊友手牌超過時 -> 隊友選
	倒數結束時客端會傳目前為止選好的(有可能回傳空陣列) -> 若為空就不理他，交由倒數結束時server自行處理
 MARKET
	發行後選擇市場更新(一定會問)
	倒數結束時客端會傳目前為止選好的(有可能回傳空陣列) -> 若為空就不理他，交由倒數結束時server自行處理
 CUBE
	抽到該阻礙卡且可選大於1
 CLARIFIED
	抽到該阻礙卡且可選大於1
*/
func (g *Game) askDrop(target int, dType string, opt ...[]int) {
	msg := transfer.AdmitAction{
		Target:   []string{g.players[target]},
		Action:   []string{DROP},
		DropType: &dType,
	}
	if opt != nil && dType == MARKET {
		msg.DropMarket = &opt
	}
	g.EventManager.Emit(transfer.DISPATCH_ADMIT_ACTION, msg)

	g.dropTicker[target] = time.NewTicker(time.Second)

	go func(pDrop int) { // receiver
		count := 0
		for {
			_ = <-g.dropTicker[pDrop].C
			if count++; count > DROP_SECOND {
				g.parseDrop(g.players[pDrop], []string{dType}, opt...)
				return
			}
			ct_msg := transfer.CountDown{
				Target:      g.players,
				Count:       DROP_SECOND - count,
				PlayerIndex: pDrop,
				CountType:   DROP,
			}
			if target != g.pNow {
				ct_msg.Target = []string{g.players[pDrop]}
			}
			g.EventManager.Emit(transfer.DISPATCH_COUNTDOWN, ct_msg)
		}
	}(target)
}

func (g *Game) parseDrop(from string, data []string, opt ...[]int) {
	pIdx := utils.FindString(g.players, from)
	if pIdx == len(g.players) {
		pIdx = g.pNow
	}
	pOfC := pIdx / 2
	cName := g.cNames[pIdx%2]
	comp := g.company[cName]

	var dType string
	if len(data) != 0 {
		dType = strings.Split(data[0], "_")[0]
		g.dropTicker[pIdx].Stop() // 讓ticker不要再送

		if dType == MARKET {
			drop := make([]int, 0, 3)
			for _, d := range data {
				dArr := strings.Split(d, "_")
				if len(dArr) > 1 {
					mIdx, _ := strconv.Atoi(strings.Split(d, "_")[1])
					drop = append(drop, mIdx)
				}
			}
			g.jungleMap.DropMarket(drop)
			g.parseAction(OVER)

		} else if dType == HAND {
			drop := make([]int, 0, 2)
			for _, d := range data {
				dArr := strings.Split(d, "_")
				if len(dArr) > 1 {
					hIdx, _ := strconv.Atoi(dArr[1])
					drop = append(drop, hIdx)
				}
			}
			sort.Sort(sort.Reverse(sort.IntSlice(drop))) // sort drop from big to small
			for _, dIdx := range drop {
				comp.DropHand(pOfC, dIdx)
			}

			l := len(comp.AllHand()[pOfC])
			for l > MAX_HAND {
				comp.DropHand(pOfC, l-1-MAX_HAND)
				l = len(comp.AllHand()[pOfC])
			}
			g.EventManager.Emit(transfer.DISPATCH_COMPANY_INFO, g.makeCompanyInfo(cName))

			if pIdx == g.pNow {
				g.drawShare()
			}

		} else {
			for _, d := range data {
				dArr := strings.Split(d, "_")
				if len(dArr) > 1 {
					pIdx, _ := strconv.Atoi(dArr[1])
					g.devMap[cName].DropSchedule(comp, pIdx)
				}
			}

			g.EventManager.Emit(transfer.DISPATCH_MAP_INFO, g.makeMapInfo())

			g.switchPlayer()
		}
	}
}

func (g *Game) endGame() {
	g.EventManager.Emit(transfer.DISPATCH_END, g.makeEndScore())

	if g.ticker != nil {
		g.ticker.Stop()
		g.ticker = nil
	}
	for i := range g.dropTicker {
		if g.dropTicker[i] != nil {
			g.dropTicker[i].Stop()
			g.dropTicker[i] = nil
		}
	}

	g.EventManager.RemoveAllListener()
	return
}

func (g *Game) makeMapInfo(draw ...transfer.DrawCard) transfer.ShareInfo {
	mapInfo := transfer.ShareInfo{
		Target:         g.players,
		Month:          g.jungleMap.Month(),
		Market:         g.jungleMap.Market(),
		BearOffice:     g.jungleMap.Office(g.cNames[0]),
		BearProgress:   g.devMap[g.cNames[0]].Schedules(),
		DeerOffice:     g.jungleMap.Office(g.cNames[1]),
		DeerProgress:   g.devMap[g.cNames[1]].Schedules(),
		PlayerPosition: g.jungleMap.Position(),
		ActionPoint:    g.pActPoint,
	}
	if len(draw) > 0 {
		mapInfo.DrawCard = &draw[0]
	}

	return mapInfo
}

func (g *Game) makeCompanyInfo(cName string, endOfM ...int) transfer.CompanyInfo {
	obj := transfer.CompanyInfo{
		Target:      g.company[cName].Players(),
		Company:     cName,
		ProductId:   g.devMap[cName].ProductIds(),
		HandCard:    g.company[cName].AllHand(),
		CheckMarket: g.jungleMap.CompanyMarket(cName),
	}
	if len(endOfM) > 0 {
		obj.EndOfMonth = &endOfM[0]
	}

	return obj
}

func (g *Game) makeEndScore() transfer.EndScore {
	sums := make([][]transfer.ProductSum, 2)
	for i, comp := range g.cNames {
		sum := g.devMap[comp].PublishSum()
		for _, s := range sum {
			for k, v := range s.Content {
				sums[i] = append(sums[i], transfer.ProductSum{Name: k, Value: v, Num: s.Num})
			}
		}
	}

	res := transfer.EndScore{
		Target:    g.players,
		BearScore: sums[0],
		DeerScore: sums[1],
	}

	return res
}
