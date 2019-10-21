package main

import (
	"fmt"
	"math/rand"
	"sort"
)

type byScore []cwAction

func (a byScore) Len() int      { return len(a) }
func (a byScore) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byScore) Less(i, j int) bool {
	distI := distance(a[i].source, a[i].target)
	distJ := distance(a[j].source, a[j].target)
	prodI := Sum(a[i].target.Production)
	prodJ := Sum(a[j].target.Production)

	return prodI/distI < prodJ/distJ
}

type cwAction struct {
	shipsToSend int
	win         int
	source      Planet
	target      Planet
	fleet0      int
	fleet1      int
	fleet2      int
}

func (g *Game) cwBestAction() (cwAction, bool) {
	myId, _ := g.getIDs()
	var actions byScore
	for _, my := range g.Planets {
		if my.OwnerID != myId {
			continue
		}
		for _, other := range g.Planets {
			if other.OwnerID == myId {
				continue
			}
			d := distance(my, other)

			attackShips, ok := g.simulateFight(my.Ships, other.getShipsAfter(d))

			if !ok {
				continue
			}

			//Check if already winning
			target := g.alreadyTargetByShips(other.Id)
			futureRes := fight(target[0], target[1], target[2], other.Ships)

			if futureRes > 0 {
				fmt.Println("already a target and winning")
				continue
			}
			f, ok := g.targetByFleet(my.Id)
			if ok {
				arriveIn := f.Eta - g.Round
				attackRes := fight(my.getShipsAfter(arriveIn)[0]-attackShips[0], my.getShipsAfter(arriveIn)[1]-attackShips[1], my.getShipsAfter(arriveIn)[2]-attackShips[2], f.Ships)
				if attackRes < 0 {
					fmt.Println("Would loose abort mission")
					continue
				}
			}

			actions = append(actions, cwAction{
				win:         1,
				shipsToSend: Sum(attackShips),
				source:      my,
				target:      other,
				fleet0:      attackShips[0],
				fleet1:      attackShips[1],
				fleet2:      attackShips[2],
			})
		}
	}

	if len(actions) == 0 {
		return cwAction{}, false
	}

	sort.Sort(actions)

	return actions[len(actions)-1], true
}

type byShips []fightRes

func (a byShips) Len() int      { return len(a) }
func (a byShips) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byShips) Less(i, j int) bool {
	return Sum(a[i].shipsToSend) < Sum(a[j].shipsToSend)
}

type fightRes struct {
	shipsToSend [3]int
}

func (g *Game) simulateFight(my [3]int, other [3]int) ([3]int, bool) {
	res := byShips{}
	for i := 1; i < 1000; i++ {
		new := my[0] - i
		if new < 0 {
			break
		}
		result := fight(new, my[1], my[2], other)
		if result > 4 {
			res = append(res, fightRes{[3]int{new, my[1], my[2]}})
		}
	}
	for i := 1; i < 1000; i++ {
		new := my[1] - i
		if new < 0 {
			break
		}
		result := fight(my[0], new, my[2], other)
		if result > 4 {
			res = append(res, fightRes{[3]int{my[0], new, my[2]}})
		}
	}
	for i := 1; i < 1000; i++ {
		new := my[2] - i
		if new < 0 {
			break
		}
		result := fight(my[0], my[1], new, other)
		if result > 4 {
			res = append(res, fightRes{[3]int{my[0], my[1], new}})
		}
	}
	for i := 1; i < 1000; i++ {
		new0 := my[0]
		new1 := my[1]
		if my[0]-i >= 0 {
			new0 -= i
		}
		if my[1]-i >= 0 {
			new1 -= i
		}
		if new0 == 0 && new1 == 0 {
			break
		}
		result := fight(new0, new1, my[2], other)
		if result > 4 {
			res = append(res, fightRes{[3]int{new0, new1, my[2]}})
		}
	}
	for i := 1; i < 1000; i++ {
		new1 := my[1]
		new2 := my[2]
		if my[1]-i >= 0 {
			new1 -= i
		}
		if my[2]-i >= 0 {
			new2 -= i
		}
		if new1 == 0 && new2 == 0 {
			break
		}
		result := fight(my[0], new1, new2, other)
		if result > 4 {
			res = append(res, fightRes{[3]int{my[0], new1, new2}})
		}
	}
	for i := 1; i < 1000; i++ {
		new0 := my[0]
		new2 := my[2]
		if my[0]-i >= 0 {
			new0 -= i
		}
		if my[2]-i >= 0 {
			new2 -= i
		}
		if new0 == 0 && new2 == 0 {
			break
		}
		result := fight(new0, my[1], new2, other)
		if result > 4 {
			res = append(res, fightRes{[3]int{new0, my[1], new2}})
		}
	}
	for i := 1; i < 1000; i++ {
		new0 := my[0]
		new1 := my[1]
		new2 := my[2]
		if my[0]-i >= 0 {
			new0 -= i
		}
		if my[1]-i >= 0 {
			new1 -= i
		}
		if my[2]-i >= 0 {
			new2 -= i
		}
		if new0 == 0 && new1 == 0 && new2 == 0 {
			break
		}
		result := fight(new0, new1, new2, other)
		if result > 4 {
			res = append(res, fightRes{[3]int{new0, new1, new2}})
		}
	}
	if len(res) == 0 {
		return [3]int{}, false
	}
	sort.Sort(res)
	return res[0].shipsToSend, true
}

func (g *Game) spray() (int, int, [3]int) {
	my := g.getOwnPlanets()
	other := g.getOtherPlanets()
	res := [3]int{}
	if len(my) == 0 || len(other) == 0 {
		return 0, 0, [3]int{}
	}
	source := rand.Intn(len(my))
	target := rand.Intn(len(other))
	ship := rand.Intn(3)
	res[ship] += 1
	return my[source].Id, other[target].Id, res
}

func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func Sum(x [3]int) int {
	return x[0] + x[1] + x[2]
}

func (g *Game) alreadyTargetByNum(other int) int {
	myid, _ := g.getIDs()
	c := 0
	for _, f := range g.Fleets {
		if f.OwnerID == myid && f.Target == other {
			c += Sum(f.Ships)
		}
	}
	return c
}

func (g *Game) alreadyTargetByShips(other int) [3]int {
	myid, _ := g.getIDs()
	c := [3]int{}
	for _, f := range g.Fleets {
		if f.OwnerID == myid && f.Target == other {
			c[0] += f.Ships[0]
			c[1] += f.Ships[1]
			c[2] += f.Ships[2]
		}
	}
	return c
}

func (g *Game) targetByFleet(other int) (Fleet, bool) {
	_, otherID := g.getIDs()
	for _, f := range g.Fleets {
		if f.OwnerID == otherID && f.Target == other {
			return f, true
		}
	}
	return Fleet{}, false
}
