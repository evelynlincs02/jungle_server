package cardpool

import (
	"jungle/server/game/pkg/utils"
	"reflect"
)

type card struct {
	id      int
	content map[string]int
}

type CardPool struct {
	deck   []int
	front  []int
	used   []int
	detail []card
}

func (cp *CardPool) GetContent(id int) map[string]int {
	// if id >= len(cp.detail) {
	// 	return nil
	// }
	// if cp.detail[id].id == id {
	// 	return cp.detail[id].content
	// } else {
	for _, d := range cp.detail {
		if d.id == id {
			return d.content
		}
	}
	// }
	return nil
}

func (cp *CardPool) GetId(content map[string]int) int {
	for _, d := range cp.detail {
		if reflect.DeepEqual(d.content, content) {
			return d.id
		}
	}
	return -1
}

func (cp *CardPool) Draw(n int, not ...int) []int {
	if n > len(cp.deck) {
		// 把 used 放回 deck。要不要再洗一下 deck 都可以反正抽的時候是隨機的
		cp.deck = append(cp.deck, cp.used...)
		cp.used = cp.used[:0]
	}

	var drawed []int
	for n > 0 {
		// 抽取 deck 內第 ri 張卡，放進 drawed，並從deck中刪除
		ri := utils.RandomInt(len(cp.deck))

		// 如果有不該抽到的牌，重抽
		if utils.FindInt(not, cp.deck[ri]) < len(not) {
			continue
		}

		drawed = append(drawed, cp.deck[ri])

		cp.deck = utils.RemoveIdx(cp.deck, ri)

		n--
	}
	cp.front = append(cp.front, drawed...)

	return drawed
}

func (cp *CardPool) DropById(id int) {
	t_i := utils.FindInt(cp.front, id)
	if t_i < len(cp.front) {
		cp.front = utils.RemoveIdx(cp.front, t_i)
		cp.used = append(cp.used, id)
	}
}

func (cp *CardPool) DetailKeys() []string {
	keys := make([]string, 0, len(cp.detail))
	d := cp.detail
	for _, c := range d {
		for k := range c.content {
			keys = append(keys, k)
		}
	}
	keys = utils.Unique(keys)
	return keys
}

func NewMarketPool() *CardPool {
	cp := new(CardPool)

	cp.deck = make([]int, 0, 35)
	for i := 0; i < 10; i++ {
		var same int
		if i < 5 {
			same = 3
		} else {
			same = 4
		}
		for same > 0 {
			same--
			cp.deck = append(cp.deck, i)
		}
	}
	cp.detail = []card{
		{
			id: 0, content: map[string]int{"shot": 5},
		},
		{
			id: 1, content: map[string]int{"chess": 5},
		},
		{
			id: 2, content: map[string]int{"adv": 5},
		},
		{
			id: 3, content: map[string]int{"sport": 5},
		},
		{
			id: 4, content: map[string]int{"puzzle": 5},
		},
		{
			id: 5, content: map[string]int{"chess": 1, "shot": 1},
		},
		{
			id: 6, content: map[string]int{"chess": 1, "puzzle": 1},
		},
		{
			id: 7, content: map[string]int{"sport": 1, "puzzle": 1},
		},
		{
			id: 8, content: map[string]int{"adv": 1, "sport": 1},
		},
		{
			id: 9, content: map[string]int{"adv": 1, "shot": 1},
		},
	}

	return cp
}

func NewProductPool() *CardPool {
	cp := new(CardPool)

	cp.deck = make([]int, 10)
	for i := range cp.deck {
		cp.deck[i] = i
	}
	cp.detail = []card{
		{
			id: 0, content: map[string]int{"chess": 4, "sport": 4},
		},
		{
			id: 1, content: map[string]int{"chess": 4, "shot": 4},
		},
		{
			id: 2, content: map[string]int{"shot": 4, "sport": 4},
		},
		{
			id: 3, content: map[string]int{"adv": 5, "puzzle": 5},
		},
		{
			id: 4, content: map[string]int{"adv": 5, "chess": 5},
		},
		{
			id: 5, content: map[string]int{"chess": 6},
		},
		{
			id: 6, content: map[string]int{"shot": 6},
		},
		{
			id: 7, content: map[string]int{"adv": 8},
		},
		{
			id: 8, content: map[string]int{"puzzle": 6},
		},
		{
			id: 9, content: map[string]int{"sport": 6},
		},
	}

	return cp
}

func NewHandPool() *CardPool {
	cp := new(CardPool)

	cp.deck = make([]int, 0, 30)
	nums := [...]int{2, 7, 6, 6, 3, 3, 3}
	for i, v := range nums {
		for ; v > 0; v-- {
			cp.deck = append(cp.deck, i)
		}
	}
	cp.detail = []card{ // map value 實際上用不到 姑且放行動點
		{
			id: 0, content: map[string]int{"idea": -1},
		},
		{
			id: 1, content: map[string]int{"clarified": -1},
		},
		{
			id: 2, content: map[string]int{"interaction": -1},
		},
		{
			id: 3, content: map[string]int{"visual": -1},
		},
		{
			id: 4, content: map[string]int{"happy": -1},
		},
		{
			id: 5, content: map[string]int{"publish": -3},
		},
		{
			id: 6, content: map[string]int{"end": 0},
		},
	}

	return cp
}

func NewChangePool() *CardPool {
	cp := new(CardPool)

	cp.deck = make([]int, 17)
	for i := range cp.deck {
		cp.deck[i] = i
	}
	cp.detail = []card{
		{
			id: 0, content: map[string]int{"bear": 0, "market": 13},
		},
		{
			id: 1, content: map[string]int{"bear": 0, "market": 12},
		},
		{
			id: 2, content: map[string]int{"bear": 1, "market": 11},
		},
		{
			id: 3, content: map[string]int{"bear": 2, "market": 8},
		},
		{
			id: 4, content: map[string]int{"bear": 3, "market": 14},
		},
		{
			id: 5, content: map[string]int{"bear": 4, "market": 15},
		},
		{
			id: 6, content: map[string]int{"bear": 5, "market": 9},
		},
		{
			id: 7, content: map[string]int{"bear": 5, "market": 10},
		},
		{
			id: 8, content: map[string]int{"deer": 0, "market": 3},
		},
		{
			id: 9, content: map[string]int{"deer": 0, "market": 2},
		},
		{
			id: 10, content: map[string]int{"deer": 1, "market": 4},
		},
		{
			id: 11, content: map[string]int{"deer": 2, "market": 7},
		},
		{
			id: 12, content: map[string]int{"deer": 3, "market": 1},
		},
		{
			id: 13, content: map[string]int{"deer": 4, "market": 0},
		},
		{
			id: 14, content: map[string]int{"deer": 5, "market": 5},
		},
		{
			id: 15, content: map[string]int{"deer": 5, "market": 6},
		},
		{
			id: 16, content: map[string]int{"obstacle": 1},
		},
	}

	return cp
}

func NewObstaclePool() *CardPool {
	cp := new(CardPool)

	cp.deck = make([]int, 17)
	nums := [...]int{2, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2}
	for i, v := range nums {
		for ; v > 0; v-- {
			cp.deck = append(cp.deck, i)
		}
	}
	cp.detail = []card{
		{
			id: 0, content: map[string]int{"p_cube": 1},
		},
		{
			id: 1, content: map[string]int{"m_sport": 3},
		},
		{
			id: 2, content: map[string]int{"m_chess": 3},
		},
		{
			id: 3, content: map[string]int{"m_shot": 3},
		},
		{
			id: 4, content: map[string]int{"m_adv": 3},
		},
		{
			id: 5, content: map[string]int{"m_puzzle": 3},
		},
		{
			id: 6, content: map[string]int{"p_clarified": 1},
		},
		{
			id: 7, content: map[string]int{"h_clarified": 4},
		},
		{
			id: 8, content: map[string]int{"h_visual": 1},
		},
		{
			id: 9, content: map[string]int{"h_interaction": 4},
		},
		{
			id: 10, content: map[string]int{"office": 3},
		},
	}

	return cp
}
