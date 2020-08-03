package company

import (
	"jungle/server/internal/cardpool"
	"jungle/server/pkg/utils"
)

type Company struct {
	players  []string
	handCard [3][]int

	handPool *cardpool.CardPool
}

func NewCompany(p []string) *Company {
	c := new(Company)

	c.players = p
	for i := range c.handCard {
		c.handCard[i] = make([]int, 0, 6)
	}
	c.handPool = cardpool.NewHandPool()

	return c
}

func (c *Company) Init() {
	for i := range c.handCard {
		c.handCard[i] = append(c.handCard[i], c.handPool.Draw(2, 6)...)
	}
}

func (c *Company) DrawHand(pIdx int) (int, bool) {
	end := 0
	newC := c.handPool.Draw(2)
	i := 0
	for _, card := range newC {
		d := c.handPool.GetContent(card)
		for k := range d {
			if k == "end" {
				c.handPool.DropById(card)
				end++
				newC = utils.RemoveIdx(newC, i)
				i-- // 刪完後順序要校正
			}
		}
		i++
	}

	c.handCard[pIdx] = append(c.handCard[pIdx], newC...)
	return end, len(c.handCard[pIdx]) > 4
}

func (c *Company) Publish(n int) {
	// 發行後放在專案那邊的設計卡要回收手牌堆
	cl := c.handPool.GetId(map[string]int{clarified.String(): -1})
	in := c.handPool.GetId(map[string]int{interaction.String(): -1})
	vi := c.handPool.GetId(map[string]int{visual.String(): -1})
	for ; n > 0; n-- {
		c.handPool.DropById(cl)
		c.handPool.DropById(in)
		c.handPool.DropById(vi)
	}
}

func (c *Company) DropHand(pIdx int, hIdx int) {
	if hIdx < len(c.handCard[pIdx]) {
		c.handPool.DropById(c.handCard[pIdx][hIdx])
		c.handCard[pIdx] = utils.RemoveIdx(c.handCard[pIdx], hIdx)
	}
}

// return needDrop
func (c *Company) GiveHand(pFrom, pTo, hIdx int) bool {
	defer func() {
		c.handCard[pFrom] = utils.RemoveIdx(c.handCard[pFrom], hIdx)
	}()
	if pTo != -1 {
		c.handCard[pTo] = append(c.handCard[pTo], c.handCard[pFrom][hIdx])
		return len(c.handCard[pTo]) > 4
	}
	return false
}

// -------------------------------- Get Functions

func (c *Company) Players() []string {
	return c.players
}

func (c *Company) AllHand() [3][]int {
	return c.handCard
}

func (c *Company) HandDetail(pi int) []string {
	hs := make([]string, 0, len(c.handCard[pi]))
	for _, val := range c.handCard[pi] {
		detail := c.handPool.GetContent(val)
		for k := range detail {
			hs = append(hs, k)
		}
	}
	return hs
}
