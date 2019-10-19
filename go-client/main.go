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
	Ships   []int `json:"ships"`
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
	Ships      []int `json:"ships"`
	Production []int `json:"production"`
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
	myId, otherId := getIDs(g.Players)

	for _, planet := range getPlanets(myId, g.Planets) {
		score += (planet.Production[0] + planet.Production[1] + planet.Production[2])
	}
	for _, planet := range getPlanets(otherId, g.Planets) {
		score -= (planet.Production[0] + planet.Production[1] + planet.Production[2])
	}

	return score
}

func (g *Game) nearestNeutral() (*Planet, *Planet) {
	myId, _ := getIDs(g.Players)
	var src, dst *Planet
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
				src = &p
				dst = &p2
			}
		}
	}
	return src, dst
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
		fmt.Println(message)
		err = json.Unmarshal([]byte(message), g)
		if err != nil {
			fmt.Printf("could not unmarshall data %v\n", err)
		}

		//myID, theirID := getIDs(g.Players)
		source, dest := g.nearestNeutral()

		if source == nil || dest == nil {
			sendNOP()
			continue
		}

		sendGameCommand(source.Id, dest.Id, source.Ships)
	}
}

func sendGameCommand(source int, target int, fleet []int) {
	command := fmt.Sprintf("send %d %d %d %d %d\n", source, target, fleet[0], fleet[1], fleet[2])
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

func getIDs(players []Player) (int, int) {
	if players[0].Itsme {
		return players[0].Id, players[1].Id
	}
	return players[1].Id, players[0].Id
}

func getPlanets(playerID int, planets []Planet) []Planet {
	res := make([]Planet, 0)
	for _, planet := range planets {
		if planet.OwnerID == playerID {
			res = append(res, planet)
		}
	}
	return res
}

func getFleets(playerID int, fleets []Fleet) []Fleet {
	res := make([]Fleet, 0)
	for _, fleet := range fleets {
		if fleet.OwnerID == playerID {
			res = append(res, fleet)
		}
	}
	return res
}
