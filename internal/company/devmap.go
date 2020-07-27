package company

import (
	"jungle/server/game/internal/cardpool"
	"jungle/server/game/pkg/utils"
	"reflect"
	"strconv"
)

type schedule uint

//go:generate stringer -type=schedule -output=devmap_string.go
const (
	none schedule = iota
	idea
	clarified
	interaction
	visual
	cube
	publish

	MAX_PROJECT  = 5
	SCHEDULE_NUM = 7
)

type project struct {
	pId      int      // Product id
	schedule schedule // 進度
	cube     int
}

func (p *project) nextSchedule() schedule {
	if p.schedule == cube && p.cube < 3 {
		return p.schedule
	} else {
		return p.schedule + 1
	}
}

func (p *project) prevSchedule() schedule {
	if p.schedule == cube && p.cube > 1 || p.schedule == none {
		return p.schedule
	} else {
		return p.schedule - 1
	}
}

type Sum struct {
	Content map[string]int
	Num     int
}

type DevMap struct {
	projects   []project
	publishSum []Sum
	publishN   int

	productPool *cardpool.CardPool
}

func NewProject() *project {
	p := new(project)

	p.pId = -1
	p.schedule = none
	p.cube = 0

	return p
}

func NewDevMap() *DevMap {
	dm := new(DevMap)

	dm.projects = make([]project, MAX_PROJECT)
	for i := range dm.projects {
		dm.projects[i] = *NewProject()
	}
	dm.publishSum = make([]Sum, 0, 10)
	dm.productPool = cardpool.NewProductPool()
	dm.publishN = 0

	// for i := 0; ; i++ { // 初始所有種類專案
	// 	d := dm.productPool.GetContent(i)
	// 	if d == nil {
	// 		break
	// 	}
	// 	dm.published = append(dm.published, Sum{d, 0})
	// }

	return dm
}

func (dm *DevMap) Init() {
	// license
	dm.projects[0].schedule = cube
	dm.projects[0].cube = 2
	// 初始project
	for i := 0; i < 3; i++ {
		dm.projects[i+1].pId = dm.productPool.Draw(1)[0]
		dm.projects[i+1].schedule = idea
	}
}

func (dm *DevMap) NewIdea(comp *Company, idx int) {
	if dm.projects[idx].schedule == none {
		dm.projects[idx].schedule = idea
		dm.projects[idx].pId = dm.productPool.Draw(1)[0]

	} else {
		dm.productPool.DropById(dm.projects[idx].pId)
		dm.projects[idx].pId = dm.productPool.Draw(1)[0]
		if dm.projects[idx].schedule > idea {
			dm.DropSchedule(comp, idx)
		}
	}
}

func (dm *DevMap) GoSchedule(idx int) {
	dm.projects[idx].schedule = dm.projects[idx].nextSchedule()
}

func (dm *DevMap) CubeUpdate(idx, cube int) {
	dm.GoSchedule(idx)
	dm.projects[idx].cube = cube
}

func (dm *DevMap) DropSchedule(comp *Company, idx int) {
	proj := &dm.projects[idx]
	oldSche := proj.schedule
	proj.schedule = proj.prevSchedule()
	if oldSche == proj.schedule && proj.cube > 0 {
		proj.cube -= 1
	} else {
		proj.cube = 0
		id := comp.handPool.GetId(map[string]int{oldSche.String(): -1})
		comp.handPool.DropById(id)
	}
}

func (dm *DevMap) Publish(pIdx []int, mNum map[string]int) (int, []string) {
	pType := make([]string, 0, 5)
	for idx := range pIdx {
		dm.publishN++
		p := dm.projects[idx+1] // 傳來的idx是不含license的(從0開始)
		detail := dm.productPool.GetContent(p.pId)
		found := false
		for i, pub := range dm.publishSum {
			found = reflect.DeepEqual(pub.Content, detail)
			if found {
				for k := range detail {
					dm.publishSum[i].Num += mNum[k]
					pType = append(pType, k)
				}
				break
			}
		}
		if !found {
			dm.publishSum = append(dm.publishSum, Sum{detail, 0})
			for k := range detail {
				dm.publishSum[len(dm.publishSum)-1].Num += mNum[k]
				pType = append(pType, k)
			}
		}
		dm.projects[idx] = *NewProject()
		dm.productPool.DropById(p.pId)
	}

	return dm.publishN, utils.Unique(pType)
}

func (dm *DevMap) CheckLisense() bool {
	if dm.projects[0].schedule == cube && dm.projects[0].cube == 3 {
		return true
	}

	return false
}

// -------------------------------- Get Functions

func (dm *DevMap) Schedules() []string {
	s := make([]string, len(dm.projects))

	for i := range dm.projects {
		s[i] = dm.projects[i].schedule.String()
		if s[i] == cube.String() {
			s[i] = s[i] + "_" + strconv.Itoa(dm.projects[i].cube)
		}
	}

	return s
}

func (dm *DevMap) ProductIds() []int {
	s := make([]int, len(dm.projects))

	for i := range dm.projects {
		s[i] = dm.projects[i].pId
	}

	return s
}

// 不排除license
func (dm *DevMap) NumOfNextSchedule(s string) int {
	sum := 0

	for _, p := range dm.projects {
		// log.Println(p.nextSchedule().String())
		if s == p.nextSchedule().String() {
			sum++
		}
	}

	return sum
}

func (dm *DevMap) PublishSum() []Sum {
	return dm.publishSum
}
