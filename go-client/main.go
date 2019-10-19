package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
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

func main() {
	fmt.Println("starting rps client")
	args := os.Args[1:]
	if len(args) != 2 {
		fmt.Println("Please provide username and password")
		return
	}

	conn, err := net.Dial("tcp", "rps.vhenne.de:6000")
	if err != nil {
		fmt.Printf("could not connect to server %v\n", err)
		return
	}
	//login
	_, err = fmt.Fprintf(conn, fmt.Sprintf("login %s %s\n", args[0], args[1]))
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
		myID := getMyID(gameData.Players)
		fmt.Println(myID)
	}
}

func sendGameCommand(conn net.Conn, source int, target int, fleet []int) {
	_, err := fmt.Fprintf(conn, fmt.Sprintf("send %d %d %d %d %d\n", source, target, fleet[0], fleet[1], fleet[2]))
	if err != nil {
		fmt.Printf("could not write to connection %v\n", err)
		return
	}
}

func getMyID(players []Player) int {
	if players[0].Itsme {
		return players[0].Id
	}
	return players[1].Id
}
