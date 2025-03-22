package main

import (
	"bufio"
	"fmt"
	"lab2/server"
	"os"
	"sync"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	var ip, port string

	fmt.Println("Choose ip to host")
	ip, _ = reader.ReadString('\n')
	ip = ip[:len(ip)-1]

	fmt.Println("Choose port to host")
	port, _ = reader.ReadString('\n')
	port = port[:len(port)-1]

	storage := server.UserStorage{Users: &sync.Map{}}

	server.StartServer(ip, port, &storage)

}
