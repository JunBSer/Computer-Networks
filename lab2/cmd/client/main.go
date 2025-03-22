package main

import (
	"bufio"
	"fmt"
	"lab2/client"
	"os"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	var ip, port string

	fmt.Println("Hello user, enter ip to connect:")
	ip, _ = reader.ReadString('\n')
	ip = ip[:len(ip)-1]

	fmt.Println("Enter port to connect:")
	port, _ = reader.ReadString('\n')
	port = port[:len(port)-1]

	client.StartClient(ip, port)
	fmt.Scan()
}
