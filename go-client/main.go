package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"os"
)

type Fleet struct {
	Id      int   `json:"id"`
	OwnerID int   `json:"owner_id"`
	Origin  int   `json:"origin"`
	Target  int   `json:"target"`
	Ships   [3]int `json:"ships"`
	Eta     int   `json:"eta"`
}

type Player struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Itsme bool   `json:"itsme"`
}

type Planet struct {
	Id         int   `json:"id"`
	OwnerID    int   `json:"owner_id"`
	X          int   `json:"x"`
	Y          int   `json:"y"`
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

type byScore []action
func (a byScore) Len() int           { return len(a) }
func (a byScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byScore) Less(i, j int) bool { return a[i].score < a[j].score }

type action struct {
	score  int
	source int
	target int
	fleet0 int
	fleet1 int
	fleet2 int
}

func (g *Game) biggestPlanet() (int, int) {
	return 0, 0
}

func (g *Game) nearestPlanet() (int, int) {
	myId, _ := g.getIDs()
	var src, dst = -1, -1
	mind := 1000000000
	for _, p := range g.Planets {
		if p.OwnerID != myId {
			continue
		}
		for _, p2 := range g.Planets {
			if p2.OwnerID == myId {
				continue
			}
			d := distance(p, p2)
			if d < mind {
				mind = d
				src = p.Id
				dst = p2.Id
			}
		}
	}
	return src, dst
}

func (p *Planet) getShipsAfter(time int) [3]int {
	return [3]int{
		p.Ships[0] + time*p.Production[0],
		p.Ships[1] + time*p.Production[1],
		p.Ships[2] + time*p.Production[2],
	}
}

func fight(_att [3]int, _def [3]int) int {
	att := [3]int{_att[0], _att[1], _att[2]}
	def := [3]int{_def[0], _def[1], _def[2]}

	for {
		s1 := att[0] + att[1] + att[2]
		s2 := def[0] + def[1] + def[2]
		if s1 == 0 || s2 == 0 {
			return s1 - s2
		}
		deltDamage := attack(att)
		recDamage := attack(def)

		for i, _ := range deltDamage {
			deltDamage[i] += def[i]
			recDamage[i] += att[i]

		}
	}
}

func attack(att [3]int) [3]int {
	def := [3]int{}
	for dType, d := range def {
		for aType, a := range att {
			if a == 0 {
				continue
			}
			mult, abs := 0, 0
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

func (g *Game) actions() (int, int) {
	myId, _ := g.getIDs()
	var actions []action
	for _, p1 := range g.Planets {
		if p1.OwnerID != myId {
			continue
		}
		for _, p2 := range g.Planets {
			if p2.OwnerID == myId {
				continue
			}
			d := distance(p1, p2)
			sc := fight(p1.Ships, p2.getShipsAfter(d))
			actions = append(actions, action{
				score:  sc,
				source: p1.Id,
				target: p2.Id,
				fleet0: p1.Ships[0],
				fleet1: p1.Ships[1],
				fleet2: p1.Ships[2],
			})
		}
	}
	return 0, 0
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
		message, err := reader.ReadString('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Printf("could not read response from server %v\n", err)
			return
		}
		if message[0] != '{' {
			fmt.Println(message)
			continue
		}
		g := &Game{}
		err = json.Unmarshal([]byte(message), g)
		if err != nil {
			fmt.Printf("could not unmarshall data %v\n", err)
		}
		source, dest := g.nearestPlanet()

		if source == -1 || dest == -1 {
			sendNOP()
			continue
		}

		srcp := g.getPlanetByID(source)
		sendGameCommand(source, dest, srcp.Ships[0], srcp.Ships[1], srcp.Ships[2])
	}
}

func sendGameCommand(source, target, fleet0, fleet1, fleet2 int) {
	command := fmt.Sprintf("send %d %d %d %d %d\n", source, target, fleet0, fleet1, fleet2)
	fmt.Println(command)
	_, err := fmt.Fprintf(conn, command)
	if err != nil {
		fmt.Printf("could not write to connection %v\n", err)
		return
	}
}

func sendNOP() {
	_, err := fmt.Fprintf(conn, "nop")
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
