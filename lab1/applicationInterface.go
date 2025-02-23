package main

import (
	"fmt"
	"net"
	"strconv"
)

func ChooseInterface(iFaces []net.Interface) (net.Interface, error) {
	fmt.Println("Choose interface")
	for i, iFace := range iFaces {
		fmt.Println(strconv.Itoa(i+1) + ")" + iFace.Name)
	}
	var chosenId int
	fmt.Println("______________________________________")
	_, err := fmt.Scanf("%d", &chosenId)
	if err != nil {
		return net.Interface{}, err
	}
	return iFaces[chosenId-1], nil
}
