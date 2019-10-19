package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
)

const (
	user = "cw-test1"
	pass = "awesomepassword"
)

type game struct {

}

func main() {
	fmt.Println("starting rps client")

	conn, err := net.Dial("tcp", "rps.vhenne.de:6000")
	if err != nil {
		fmt.Printf("could not connect to server %v", err)
		return
	}
	//login
	fmt.Fprintf(conn, fmt.Sprintf("login %s %s\n", user, pass))
	for {
		//read
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			fmt.Printf("could not read response from server %v", err)
			return
		}
		fmt.Print("Message from server: " + message)
	}
}
