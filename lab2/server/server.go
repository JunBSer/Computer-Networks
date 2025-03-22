package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"lab2/client"
	"lab2/models"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
)

const network = "tcp"
const maxUserCnt = 10

func GenerateID(lastValue int32) int32 {
	return lastValue + 1
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

func CloseConnection(usr *models.User, storage Storage, connCnt *atomic.Int32) {
	var err error
	if usr.Conn != nil {
		err = usr.Conn.Close()
	}
	if err != nil {
		log.Fatalf("Error closing connection: %v", err)
	}
	log.Println("Connection closed.", usr.UID, usr.Name)
	storage.RemoveUser(usr.UID)
	connCnt.Add(-1)
}

func HandleConnection(ctx context.Context, conn net.Conn, id int32, storage Storage, wg *sync.WaitGroup, connCnt *atomic.Int32) {
	var usr *models.User
	msg := &models.Message{}

	defer (*wg).Done()

	// Can be upgraded by adjusting connection handshake
	data, err := models.GetMessage(conn)
	if err != nil {
		log.Printf("Can not get name from user %d", id)
		return
	}

	msg, err = models.ParseMsg(data)
	if err != nil {
		log.Printf("Can not parse message from user %d", id)
		return
	}

	var usrName string
	err = json.Unmarshal(msg.Body, &usrName)
	if err != nil {
		log.Printf("Username of %d was not parsed", id)
		return
	}

	usr = client.CreateUser(id, conn, usrName)

	defer CloseConnection(usr, storage, connCnt)

	storage.AddUser(usr)

	jId, _ := json.Marshal(id)
	msg.Body = json.RawMessage(jId)
	msg.ReceiverId = id
	msg.SenderID = -1

	if err = models.SendMessage(conn, msg); err != nil {
		log.Printf("Can not send message with id to user %d", id)
	}

	log.Println("Accepted connection:", id, usrName)

	for {
		select {
		case <-ctx.Done():
			log.Println("The user has been disconnected by shutting the server", usrName, usr.UID)
			return

		default:
			data, err = models.GetMessage(conn)
			if err != nil {
				log.Printf("Can not get message from user %d name: %s", id, usrName)
				if errors.Is(err, io.EOF) {
					log.Println("Connection closed.")

				}
				return
			}

			msg, err = models.ParseMsg(data)
			if err != nil {
				log.Printf("Can not parse message from user %d name: %s", id, usrName)
				continue
			}

			if msg.ReceiverId == -1 {
				msg.SenderID = -1
				msg.ReceiverId = usr.UID
				msg.Body, err = json.Marshal(storage.GetAllUsersData())
				if err = models.SendMessage(conn, msg); err != nil {
					log.Printf("Can not send message with users db info to user %d name: %s", id, usrName)
				}
			} else {
				receiver := msg.ReceiverId

				resConn := storage.GetUserById(receiver)
				if resConn == nil {
					continue
				}

				if err = models.SendMessage(storage.GetUserById(receiver).Conn, msg); err != nil {
					log.Printf("Can not send message to user %d -- from user %d name %s", receiver, id, usrName)
				}

			}
		}
	}
}

func SendUsrList(usr *models.User, storage Storage) {
	users := storage.GetAllUsersData()
	if len(users) == 0 {
		return
	}
	jUsers, err := json.Marshal(users)
	msg := &models.Message{
		ReceiverId: usr.UID,
		SenderID:   -1,
		Body:       json.RawMessage(jUsers),
	}

	err = models.SendMessage(usr.Conn, msg)
	if err != nil {
		log.Printf("Can not send user list to user %d name: %s", usr.UID, usr.Name)
	}
}

func SendUsrListToAll(storage Storage) {
	users := storage.GetAllUsersData()
	if len(users) == 0 {
		return
	}

	for _, usr := range users {
		SendUsrList(storage.GetUserById(usr.UID), storage)
	}
}

func StartServer(addr, port string, storage Storage) {
	port = strings.TrimSpace(port)
	addr = strings.TrimSpace(addr)
	fullAddr := fmt.Sprintf("%s:%s", addr, port)

	log.Println("Starting server...\nAddr: " + fullAddr)

	l, err := net.Listen(network, fullAddr)
	if err != nil {
		log.Printf("Cant listen on the addr %s <------> %v", fullAddr, err)
		return
	}

	var baseId int32 = 0
	var connCnt atomic.Int32
	connCnt.Store(0)

	ctx := SetupGracefulShutdown()
	var wg sync.WaitGroup

	for {
		select {
		case <-ctx.Done():
			log.Println("The server is shutting down...")
			return
		default:
			if connCnt.Load() == maxUserCnt {
				continue
			}

			conn, err := l.Accept()
			if err != nil {
				log.Println("Error accepting: ", err.Error())
			}

			connCnt.Add(1)

			newId := GenerateID(baseId)

			wg.Add(1)
			go HandleConnection(ctx, conn, newId, storage, &wg, &connCnt)

			baseId = newId
			//SendUsrListToAll(storage)
		}

	}

}
