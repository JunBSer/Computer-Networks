package client

import (
	"fmt"
	"log"
	"strings"
)

const (
	refreshConn = iota
	chooseComp
	outputInfo
)

var displayMethods [3]func(client *Client) = [3]func(client *Client){RefreshClUsers, ChangeChatCompanion, OutputAvailableUsers}

func ProcessUserInput(str string, client *Client) bool {
	fl := false
	str = strings.TrimSpace(str)
	str = strings.Trim(str, "\r")
	//str = strings.Trim(str, "\n ")
	switch str {

	case "/0":
		displayMethods[refreshConn](client)
		fl = true
	case "/1":
		displayMethods[chooseComp](client)
		fl = true
	case "/2":
		displayMethods[outputInfo](client)
	}
	return fl

}

func ChangeChatCompanion(cl *Client) {
	fmt.Println("Choose your companion for chatting")
	cl.connMu.Lock()
	for i, val := range cl.connUsers {
		fmt.Printf("%d-%v\n", i, val)
	}
	cl.connMu.Unlock()
	fmt.Println()

	var currComp int32 = -1

	_, err := fmt.Scan(&currComp)
	if err != nil {
		log.Println("Error reading companion id")
		return
	}

	cl.compMu.Lock()
	*cl.currComp = currComp
	defer cl.compMu.Unlock()
	fmt.Println("Companion successfully changed")
}

func RefreshClUsers(cl *Client) {
	log.Println("Refreshing client user list")
	fmt.Println("Next you need to choose your companion again")

	cl.compMu.Lock()
	*cl.currComp = -1
	defer cl.compMu.Unlock()

}

func OutputAvailableUsers(cl *Client) {
	fmt.Println("Available users:")
	cl.connMu.Lock()
	for i, val := range cl.connUsers {
		fmt.Printf("%d-%v\n", i, val)
	}
	cl.connMu.Unlock()
}
