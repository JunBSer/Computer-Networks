package models

import "net"

type User struct {
	UserData
	Conn net.Conn
}

type UserData struct {
	UID  int32  `json:"UID"`
	Name string `json:"Name"`
}
