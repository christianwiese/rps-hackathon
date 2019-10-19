package main

import (
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
	if a[i].win == a[j].win {
		if prodI == prodJ {
			return distI > distJ
		}
		return prodI < prodJ
	}
	return a[i].win < a[j].win
}

type cwAction struct {
	win    int
	source Planet
	target Planet
	fleet0 int
	fleet1 int
	fleet2 int
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

			result := fight(my.Ships, other.getShipsAfter(d))
			if result < 0 {
				continue
			}

			actions = append(actions, cwAction{
				win:    1,
				source: my,
				target: other,
				fleet0: my.Ships[0],
				fleet1: my.Ships[1],
				fleet2: my.Ships[2],
			})
		}
	}

	if len(actions) == 0 {
		return cwAction{}, false
	}

	sort.Sort(actions)

	return actions[len(actions)-1], true
}

//func cwFight(_att [3]int, _def [3]int) (int, [3]int) {
//	att := [3]float64{float64(_att[0]), float64(_att[1]), float64(_att[2])}
//	def := [3]float64{float64(_def[0]), float64(_def[1]), float64(_def[2])}
//
//	for {
//		s1 := att[0] + att[1] + att[2]
//		s2 := def[0] + def[1] + def[2]
//		//fmt.Println("fighting...", s1, s2)
//		if s1 <= 0 || s2 <= 0 {
//			return int(s1 - s2), att
//		}
//		deltDamage := attack(att)
//		recDamage := attack(def)
//
//		for i, _ := range deltDamage {
//			def[i] = math.Max(def[i]+deltDamage[i], 0)
//			att[i] = math.Max(att[i]+recDamage[i], 0)
//		}
//	}
//}

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
