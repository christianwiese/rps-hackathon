package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"os"
	"strings"
)

type Fleet struct {
	Id      int    `json:"id"`
	OwnerID int    `json:"owner_id"`
	Origin  int    `json:"origin"`
	Target  int    `json:"target"`
	Ships   [3]int `json:"ships"`
	Eta     int    `json:"eta"`
}

type Player struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Itsme bool   `json:"itsme"`
}

type Planet struct {
	Id         int    `json:"id"`
	OwnerID    int    `json:"owner_id"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	Ships      [3]int `json:"ships"`
	Production [3]int `json:"production"`
}

type Game struct {
	GameOver  bool     `json:"game_over"`
	Winner    int      `json:"winner"`
	Round     int      `json:"round"`
	MaxRounds int      `json:"max_rounds"`
	Fleets    []Fleet  `json:"fleets"`
	Players   []Player `json:"players"`
	Planets   []Planet `json:"planets"`
}

var conn net.Conn

func (g *Game) score() int {
	if g.GameOver {
		return 0
	}
	score := 0
	myId, otherId := g.getIDs()

	for _, planet := range g.getPlanets(myId) {
		score += (planet.Production[0] + planet.Production[1] + planet.Production[2])
	}
	for _, planet := range g.getPlanets(otherId) {
		score -= (planet.Production[0] + planet.Production[1] + planet.Production[2])
	}

	return score
}

type action struct {
	score  int
	source int
	target int
	fleet0 int
	fleet1 int
	fleet2 int
}

func (g *Game) nearestPlanet() (int, int) {
	myId, _ := g.getIDs()
	biggest, _ := g.biggestOwnPlanet()
	var src, dst = -1, -1
	mind := 1000000000

	for _, p2 := range g.Planets {
		if p2.OwnerID == myId {
			continue
		}
		d := distance(*g.getPlanetByID(biggest), p2)
		if d < mind {
			mind = d
			src = biggest
			dst = p2.Id
		}
	}
	return src, dst
}

func (p *Planet) getShipsAfter(time int) [3]int {
	if p.OwnerID == 0 {
		return p.Ships
	}
	return [3]int{
		p.Ships[0] + time*p.Production[0],
		p.Ships[1] + time*p.Production[1],
		p.Ships[2] + time*p.Production[2],
	}
}

func fight(_att [3]int, _def [3]int) int {
	att := [3]float64{float64(_att[0]), float64(_att[1]), float64(_att[2])}
	def := [3]float64{float64(_def[0]), float64(_def[1]), float64(_def[2])}

	sc := 0.0

	for {
		s1 := att[0] + att[1] + att[2]
		s2 := def[0] + def[1] + def[2]
		//fmt.Println("fighting...", s1, s2)
		if s1 <= 0 || s2 <= 0 {
			return int(s1 - s2)
		}
		deltDamage := attack(att)
		recDamage := attack(def)

		sc += deltDamage[0] + deltDamage[1] + deltDamage[2] - recDamage[0] - recDamage[1] - recDamage[2]

		for i, _ := range deltDamage {
			def[i] = math.Max(def[i]+deltDamage[i], 0)
			att[i] = math.Max(att[i]+recDamage[i], 0)
		}
	}
}

func attack(att [3]float64) [3]float64 {
	def := [3]float64{}
	for dType, d := range def {
		for aType, a := range att {
			if a == 0 {
				continue
			}
			mult, abs := 0.0, 0.0
			if aType == dType {
				mult = 0.1
				abs = 1
			} else if (dType-aType+3)%3 == 1 {
				mult = 0.25
				abs = 2
			} else if (dType-aType+3)%3 == 2 {
				mult = 0.01
				abs = 1
			} else {
				panic("impossible!!")
			}
			if a*mult < abs {
				d -= abs
			} else {
				d -= a * mult
			}
		}
		def[dType] = d
	}
	return def
}

func (g *Game) bestAction() action {
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
			c := g.alreadySent(my.Id, other.Id)
			d := distance(my, other)
			after := other.getShipsAfter(d)
			val := (other.Production[0]) + (other.Production[1]) + (other.Production[2])

			send := [3]int{}
			for s0 := 1; s0 < 1000; s0++ {
				r := rand.Intn(after[0] + after[1] + after[2] + 1)
				if r <= after[0] {
					send[2] += rand.Intn(10)
				} else if r <= after[0]+after[1] {
					send[0] += rand.Intn(10)
				} else {
					send[1] += rand.Intn(10)
				}
				sc := fight(send, after)
				if sc > 0 {
					//fmt.Println("score", sc)
					actions = append(actions, action{
						score:  val - d - c + rand.Intn(2),
						source: my.Id,
						target: other.Id,
						fleet0: send[0] + 1,
						fleet1: send[1] + 1,
						fleet2: send[2] + 1,
					})
					break
				}
			}
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

func distance(p1 Planet, p2 Planet) int {
	var dx = float64(p1.X - p2.X)
	var dy = float64(p1.Y - p2.Y)
	return int(math.Sqrt(dx*dx+dy*dy) + 0.9999)
}

func main() {
	fmt.Println("starting rps client")
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Println("Please provide username and password")
		return
	}

	var game int
	for {
		game++
		fmt.Printf("starting game %d\n", game)
		var err error
		conn, err = net.Dial("tcp", "rps.vhenne.de:6000")
		if err != nil {
			fmt.Printf("could not connect to server %v\n", err)
			return
		}
		//login
		_, err = fmt.Fprintf(conn, "login %s %s\n", args[0], args[1])
		if err != nil {
			fmt.Printf("could not write to connection %v\n", err)
			return
		}

		reader := bufio.NewReader(conn)
		for {
			//read
			m, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				fmt.Printf("could not read response from server %v\n", err)
				return
			}
			if strings.HasPrefix(m, "calculating round") || strings.HasPrefix(m, "command received. waiting for other player...") || strings.HasPrefix(m, "waiting for you command") {
				continue
			}

			if m[0] != '{' {
				fmt.Println(m)
				continue
			}
			g := &Game{}
			err = json.Unmarshal([]byte(m), g)
			if err != nil {
				fmt.Printf("could not unmarshall data %v\n", err)
			}

			action, ok := g.cwBestAction()

			if !ok {
				sendNOP()
				continue
			}

			sendGameCommand(action.source.Id, action.target.Id, action.fleet0, action.fleet1, action.fleet2)
		}
	}
}

func sendGameCommand(source, target, fleet0, fleet1, fleet2 int) {
	if fleet0 < 0 {
		fleet0 = 0
	}
	if fleet1 < 0 {
		fleet1 = 0
	}
	if fleet2 < 0 {
		fleet2 = 0
	}
	command := fmt.Sprintf("send %d %d %d %d %d\n", source, target, fleet0, fleet1, fleet2)
	fmt.Print(command)
	_, err := fmt.Fprintf(conn, command)
	if err != nil {
		fmt.Printf("could not write to connection %v\n", err)
		return
	}
}

func sendNOP() {
	fmt.Println("send nop")
	_, err := fmt.Fprint(conn, "nop\n")
	if err != nil {
		fmt.Printf("could not write to connection %v\n", err)
		return
	}
}

func (g *Game) getIDs() (int, int) {
	if g.Players[0].Itsme {
		return g.Players[0].Id, g.Players[1].Id
	}
	return g.Players[1].Id, g.Players[0].Id
}

func (g *Game) getOwnPlanets() []Planet {
	me, _ := g.getIDs()
	return g.getPlanets(me)
}

func (g *Game) biggestOwnPlanet() (int, int) {
	var id int
	var fleet int
	ownPlanets := g.getOwnPlanets()
	for _, p := range ownPlanets {
		sum := p.Ships[0] + p.Ships[1] + p.Ships[2]
		if sum > fleet {
			id = p.Id
			fleet = sum
		}
	}
	return id, fleet
}

func (g *Game) getPlanets(playerID int) []Planet {
	res := make([]Planet, 0)
	for _, planet := range g.Planets {
		if planet.OwnerID == playerID {
			res = append(res, planet)
		}
	}
	return res
}

func (g *Game) getFleets(playerID int) []Fleet {
	res := make([]Fleet, 0)
	for _, fleet := range g.Fleets {
		if fleet.OwnerID == playerID {
			res = append(res, fleet)
		}
	}
	return res
}

func (g *Game) getPlanetByID(planetID int) *Planet {
	if planetID == -1 {
		return nil
	}
	var res Planet
	for _, planet := range g.Planets {
		if planet.Id == planetID {
			res = planet
		}
	}
	return &res
}

func (g *Game) alreadySent(my int, other int) int {
	myid, _ := g.getIDs()
	c := 0
	for _, f := range g.Fleets {
		if f.OwnerID == myid && f.Origin == my && f.Target == other {
			c++
		}
	}
	return c
}

func (g *Game) getMyFleetsForTarget(target int) []*Fleet {
	var fleet []*Fleet
	me, _ := g.getIDs()
	fleets := g.getFleets(me)
	for _, f := range fleets {
		fleet = append(fleet, &f)
	}
	return fleet
}
