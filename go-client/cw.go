package main

func (g *Game) cwBestAction() action {
	myId, _ := g.getIDs()
	var actions []action
	for _, my := range g.Planets {
		if my.OwnerID != myId {
			continue
		}
		for _, other := range g.Planets {
			if other.OwnerID == myId {
				continue
			}
			d := distance(my, other)

			var score int
			//sc, _ := cwFight(my.Ships, other.getShipsAfter(d))
			//fmt.Println("score", sc)
			win1 := my.Ships[0] - other.getShipsAfter(d)[1]
			win2 := my.Ships[1] - other.getShipsAfter(d)[2]
			win3 := my.Ships[2] - other.getShipsAfter(d)[0]

			if win1 > 0 {
				score += win1
			}
			if win2 > 0 {
				score += win2
			}
			if win3 > 0 {
				score += win3
			}

			actions = append(actions, action{
				score:  score,
				source: my.Id,
				target: other.Id,
				fleet0: win1,
				fleet1: win2,
				fleet2: win3,
			})
		}
	}
	action := action{score: -1000000000}
	for _, a := range actions {
		if a.score > action.score {
			action = a
		}
	}
	return action
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
