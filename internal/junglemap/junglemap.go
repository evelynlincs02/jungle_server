package junglemap

import (
	"jungle/server/game/internal/cardpool"
	"jungle/server/game/pkg/utils"
	"strconv"
	"strings"
)

type companyname int

//go:generate stringer -type=companyname -output=junglemap_string.go
const (
	bear companyname = iota
	deer
)

type JungleMap struct {
	companys [2]string

	month        int
	marketGrid   [16]int
	shareFront   [16]bool
	companyFront map[string][]int
	officeHappy  map[string][]bool
	position     [6]string

	marketPool *cardpool.CardPool
}

func NewJungleMap() *JungleMap {
	j := new(JungleMap)

	j.companys = [2]string{bear.String(), deer.String()}

	j.month = 1
	j.companyFront = make(map[string][]int, 2)
	j.officeHappy = make(map[string][]bool, 2)
	for _, c := range j.companys {
		j.companyFront[c] = []int{}
		j.officeHappy[c] = make([]bool, 6)
	}
	j.marketPool = cardpool.NewMarketPool()

	return j
}

func (j *JungleMap) Init() {
	for i := 0; i < 6; i++ {
		for _, c := range j.companys {
			j.officeHappy[c][i] = true
		}
		j.position[i] = "Unknown"
	}
	// 抽市場
	draw := j.marketPool.Draw(16)
	copy(j.marketGrid[:], draw)
	// 翻正初始八張
	m := make([]bool, 8)
	for i := range m {
		m[i] = true
	}
	copy(j.shareFront[4:12], m)
}

func (j *JungleMap) Change(detail map[string]int, cNow string) string {
	if _, ok := detail["obstacle"]; ok {
		return cNow
	}

	var obstacleTarget string = ""

	// 市場變動
	for k, v := range detail {
		if k == "market" { // market
			mIdx := detail["market"]
			if !j.shareFront[mIdx] { // 反面，翻正
				j.shareFront[mIdx] = true
			} else { // 正面，抽新的
				j.marketPool.DropById(j.marketGrid[mIdx])
				newC := j.marketPool.Draw(1)
				j.marketGrid[mIdx] = newC[0]
				j.shareFront[mIdx] = false
			}

			for _, c := range j.companys {
				// 市場改了，玩家有看過的話要刪掉
				if f_i := utils.FindInt(j.companyFront[c], mIdx); f_i < len(j.companyFront[c]) {
					j.companyFront[c] = utils.RemoveIdx(j.companyFront[c], f_i)
				}
			}
		} else { // office
			// 辦公室人員轉為不開心
			if j.officeHappy[k][v] {
				j.officeHappy[k][v] = false
			} else { // 本來就不開心，該辦公室抽阻礙
				obstacleTarget = k
			}
		}
	}

	return obstacleTarget
}

func (j *JungleMap) DropMarket(target []int) {
	newC := j.marketPool.Draw(len(target))
	for i, mIdx := range target {
		j.marketPool.DropById(j.marketGrid[mIdx])
		j.marketGrid[mIdx] = newC[i]
		j.shareFront[mIdx] = false

		for _, c := range j.companys {
			// 市場改了，玩家有看過的話要刪掉
			if f_i := utils.FindInt(j.companyFront[c], mIdx); f_i < len(j.companyFront[c]) {
				j.companyFront[c] = utils.RemoveIdx(j.companyFront[c], f_i)
			}
		}
	}
}

func (j *JungleMap) MakeHappy(c string, happy bool, targets ...int) {
	for _, t := range targets {
		j.officeHappy[c][t] = happy
	}
}

func (j *JungleMap) GoPos(pIdx int, pos string) {
	j.position[pIdx] = pos
}

func (j *JungleMap) AddMonth(n int) {
	j.month += n
}

func (j *JungleMap) PosCanHappy(c string, pIdx int) bool {
	pos := j.position[pIdx]
	if pos == "Unknown" {
		return false
	}

	posArr := strings.Split(pos, "_")
	location := posArr[0]
	lIdx, _ := strconv.Atoi(posArr[1])

	if location == "office" {
		switch lIdx {
		case 0:
			return !j.officeHappy[c][0] || !j.officeHappy[c][1] || !j.officeHappy[c][3]
		case 1:
			return !j.officeHappy[c][0] || !j.officeHappy[c][1] || !j.officeHappy[c][2] || !j.officeHappy[c][4]
		case 2:
			return !j.officeHappy[c][1] || !j.officeHappy[c][2] || !j.officeHappy[c][5]
		case 3:
			return !j.officeHappy[c][0] || !j.officeHappy[c][3] || !j.officeHappy[c][4]
		case 4:
			return !j.officeHappy[c][1] || !j.officeHappy[c][3] || !j.officeHappy[c][4] || !j.officeHappy[c][5]
		case 5:
			return !j.officeHappy[c][2] || !j.officeHappy[c][4] || !j.officeHappy[c][5]
		}
	} else { // location ==  "market"
		if c == j.companys[0] {
			switch lIdx {
			case 0:
				return !j.officeHappy[c][5]
			case 1:
				return !j.officeHappy[c][4]
			case 2:
				return !j.officeHappy[c][3]
			}
		} else {
			switch lIdx {
			case 15:
				return !j.officeHappy[c][5]
			case 14:
				return !j.officeHappy[c][4]
			case 13:
				return !j.officeHappy[c][3]
			}
		}

	}

	return false
}

func (j *JungleMap) PosCanMarket(c string, pIdx int) bool {
	pos := j.position[pIdx]
	if pos == "Unknown" {
		return false
	}

	posArr := strings.Split(pos, "_")
	location := posArr[0]
	lIdx, _ := strconv.Atoi(posArr[1])

	if location == "office" {
		if lIdx < 3 {
			return false
		} else {
			switch lIdx {
			case 3:
				if c == j.companys[0] {
					return !j.shareFront[2] && utils.FindInt(j.companyFront[c], 2) == len(j.companyFront[c])
				} else {
					return !j.shareFront[13] && utils.FindInt(j.companyFront[c], 13) == len(j.companyFront[c])
				}
			case 4:
				if c == j.companys[0] {
					return !j.shareFront[1] && utils.FindInt(j.companyFront[c], 1) == len(j.companyFront[c])
				} else {
					return !j.shareFront[14] && utils.FindInt(j.companyFront[c], 14) == len(j.companyFront[c])
				}
			case 5:
				if c == j.companys[0] {
					return !j.shareFront[0] && utils.FindInt(j.companyFront[c], 0) == len(j.companyFront[c])
				} else {
					return !j.shareFront[15] && utils.FindInt(j.companyFront[c], 15) == len(j.companyFront[c])
				}
			}
		}
	} else { // location ==  "market"
		l := lIdx - 4
		r := lIdx + 4
		u := lIdx + 1
		if lIdx%4 == 3 {
			u = -1
		}
		d := lIdx - 1
		if lIdx%4 == 0 {
			d = -1
		}

		candidate := [...]int{l, r, u, d, lIdx}

		for _, i := range candidate {
			if i >= 0 && i < 16 {
				if !j.shareFront[i] && utils.FindInt(j.companyFront[c], i) == len(j.companyFront[c]) {
					return true
				}
			}
		}
	}

	return false
}

func (j *JungleMap) CheckSamePos(pIdx int) bool {
	myPos := j.position[pIdx]
	if myPos != "Unknown" {
		for i := 0; i < 3; i++ {
			partner := i*2 + pIdx%2
			if partner != pIdx && myPos == j.position[partner] {
				return true
			}
		}
	}
	return false
}

func (j *JungleMap) CheckMarket(c string, mIdx int) {
	j.companyFront[c] = append(j.companyFront[c], mIdx)
}

// -------------------------------- Get Functions

func (j *JungleMap) Month() int {
	return j.month
}

func (j *JungleMap) Market() [16]int {
	market := new([16]int)
	for i, g := range j.marketGrid {
		if j.shareFront[i] {
			market[i] = g
		} else {
			market[i] = -1
		}
	}
	return *market
}

func (j *JungleMap) Office(c string) []bool {
	return j.officeHappy[c]
}

func (j *JungleMap) Position() [6]string {
	return j.position
}

func (j *JungleMap) CompanyMarket(c string) []string {
	r := make([]string, 0)
	for _, v := range j.companyFront[c] {
		r = append(r, strconv.Itoa(v)+"_"+strconv.Itoa(j.marketGrid[v]))
	}
	return r
}

func (j *JungleMap) NumOfHappy(c string, targets ...int) int {
	n := 0
	if targets == nil || len(targets) == 0 {
		targets = []int{0, 1, 2, 3, 4, 5}
	}
	for _, v := range targets {
		if j.officeHappy[c][v] {
			n++
		}
	}
	return n
}

func (j *JungleMap) MarketSum(ad ...int) map[string]int {
	sum := make(map[string]int, len(j.marketPool.DetailKeys()))
	for _, m := range j.Market() {
		if m != -1 {
			c := j.marketPool.GetContent(m)
			for k, v := range c {
				sum[k] += v
			}
		}
	}
	return sum
}

func (j *JungleMap) DropOption(target []string) [][]int {
	ans := make([][]int, 2)

	for _, t := range target {
		for i, mId := range j.Market() {
			if mId == -1 {
				continue
			}
			m := j.marketPool.GetContent(mId)
			for k, v := range m {
				if k == t {
					if v == 5 {
						ans[0] = append(ans[0], i)
					} else if v == 1 {
						ans[1] = append(ans[1], i)
					}
				}
			}
		}
	}

	return ans
}
