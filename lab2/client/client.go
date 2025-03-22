package client

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"lab2/models"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Client struct {
	connUsers map[int32]string
	currComp  *int32
	compMu    sync.Mutex
	connMu    sync.Mutex
}

func CreateUser(id int32, conn net.Conn, Name string) *models.User {
	return &models.User{
		UserData: models.UserData{
			UID:  id,
			Name: Name,
		},
		Conn: conn,
	}
}

func SetupGracefulShutdown() context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalChan
		log.Println("Service is shutting down...")
		cancel()
	}()

	return ctx
}

func Output(msg *models.Message, cl *Client) error {
	var text string
	err := json.Unmarshal(msg.Body, &text)
	if err != nil {
		log.Print("Error to complete string conversion from msg")
	}
	cl.connMu.Lock()
	defer cl.connMu.Unlock()
	usr := cl.connUsers[msg.SenderID]

	_, err = fmt.Printf("%s > %s\n", usr, text)
	return err
}

func ReadConn(ctx context.Context, conn net.Conn, cl *Client) {
	var data []byte
	var msg *models.Message
	var err error

	for {
		select {
		case <-ctx.Done():
			log.Printf("Read connection is closed")
			return

		default:
			data, err = models.GetMessage(conn)
			if err != nil {

				if errors.Is(err, io.EOF) {
					log.Printf("Error to read message from conn %v", err)
				}
				log.Printf("Error getting message: %s\n", err)
				return
			}

			msg, err = models.ParseMsg(data)
			if err != nil {
				log.Printf("Error parsing message(usrid): %s\n", err)
			}

			if msg.SenderID == -1 {
				var users []models.UserData
				err = json.Unmarshal(msg.Body, &users)
				if err != nil {
					log.Printf("Error parsing message with userlist (usrid): %s\n", err)
					continue
				}

				cl.connMu.Lock()
				for _, u := range users {
					if _, ok := cl.connUsers[u.UID]; !ok {
						cl.connUsers[u.UID] = u.Name
					} else {
						delete(cl.connUsers, u.UID)
					}
				}
				cl.connMu.Unlock()
				continue
			}

			err = Output(msg, cl)
			if err != nil {
				log.Printf("Error outputting message(usrid): %s\n", err)
			}
		}
	}
}

func ReadConsole(ctx context.Context, consChan chan string, cl *Client) {
	reader := bufio.NewReader(os.Stdin)
	for {
		select {
		case <-ctx.Done():
			log.Printf("Read console is closed")
			return

		default:

			fmt.Println("Enter /0 to refresh userList")
			fmt.Println("Enter /1 to change companion")
			fmt.Println("Enter /2 to output available users")
			fmt.Println("Enter '-' to read from console\n")

			for {
				var str string
				for {
					line, err := reader.ReadString('\n')
					line = strings.TrimSpace(line)

					if err != nil {
						if err.Error() == "EOF" {
							break
						}
						fmt.Println("Error while reading the console: ", err)
						return
					}
					if line == "-" {
						break
					}
					str += line
				}

				if len(str) != 0 {
					if !ProcessUserInput(str, cl) || str == "/0\n" {
						consChan <- str
					}
				}

			}
		}
	}
}

func StartClient(addr, port string) {
	usr := &models.User{}
	msg := &models.Message{}

	cl := Client{
		connUsers: make(map[int32]string),
		currComp:  new(int32),
		compMu:    sync.Mutex{},
		connMu:    sync.Mutex{},
	}

	port = strings.TrimSpace(port)
	addr = strings.TrimSpace(addr)
	fullAddr := fmt.Sprintf("%s:%s", addr, port)

	log.Println("Starting client...\nAddr: " + fullAddr)

	conn, err := net.Dial("tcp", fullAddr)
	if err != nil {
		log.Printf("Error connecting to server: %s\n", err)
		return
	}

	defer conn.Close()

	var usrName string
	fmt.Printf("Input name: \n")
	_, _ = fmt.Scanln(&usrName)
	usr.Name = usrName

	jName, err := json.Marshal(usrName)
	if err = models.SendMessage(conn, &models.Message{SenderID: 0, ReceiverId: -1, Body: json.RawMessage(jName)}); err != nil {
		log.Printf("Error sending message: %s\n", err)
		return
	}

	var data []byte

	data, err = models.GetMessage(conn)
	if err != nil {
		log.Printf("Error getting message(usrid): %s\n", err)
		return
	}

	msg, err = models.ParseMsg(data)
	if err != nil {
		log.Printf("Error parsing message(usrid): %s\n", err)
		return
	}

	err = json.Unmarshal(msg.Body, &usr.UID)
	if err != nil {
		log.Printf("Error getting usrid from msg: %s\n", err)
		return
	}
	log.Printf("You was succesfully registred by server UsrId: %d\n", usr.UID)

	usr.Conn = conn

	consChan := make(chan string)
	ctx := SetupGracefulShutdown()
	*cl.currComp = -1

	go ReadConn(ctx, conn, &cl)
	go ReadConsole(ctx, consChan, &cl)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Connection is closed")
			return

		case str, ok := <-consChan:
			if !ok {
				log.Printf("Error reading from channel: %s\n", err)
				continue
			}
			jStr, err := json.Marshal(str)
			msg.Body = json.RawMessage(jStr)
			msg.SenderID = usr.UID

			cl.compMu.Lock()
			msg.ReceiverId = *cl.currComp
			cl.compMu.Unlock()

			if err = models.SendMessage(conn, msg); err != nil {
				log.Printf("Error sending message: %s\n", err)
			}

		default:
			time.Sleep(2 * time.Second)
		}

	}
}
