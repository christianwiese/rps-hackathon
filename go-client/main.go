package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

const (
	user = "cw-test2"
	pass = "awesomepassword"
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

func main() {
	fmt.Println("starting rps client")

	conn, err := net.Dial("tcp", "rps.vhenne.de:6000")
	if err != nil {
		fmt.Printf("could not connect to server %v\n", err)
		return
	}
	//login
	_, err = fmt.Fprintf(conn, fmt.Sprintf("login %s %s\n", user, pass))
	if err != nil {
		fmt.Printf("could not write to connection %v\n", err)
		return
	}
	for {
		//read
		message, err := bufio.NewReader(conn).ReadString('\n')
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
		gameData := &Game{}
		fmt.Println(message)
		err = json.Unmarshal([]byte(message), gameData)
		if err != nil {
			fmt.Printf("could not unmarshall data %v\n", err)
		}
		//fmt.Printf("game state: %+v", gameData)
		////////////////
		myID, theirID := getIDs(gameData.Players)
		fmt.Println(myID, theirID)
	}
}

func sendGameCommand(conn net.Conn, source int, target int, fleet []int) {
	_, err := fmt.Fprintf(conn, fmt.Sprintf("send %d %d %d %d %d\n", source, target, fleet[0], fleet[1], fleet[2]))
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
