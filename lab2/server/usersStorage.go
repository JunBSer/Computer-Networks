package server

import (
	"lab2/models"
	"sync"
)

type Storage interface {
	GetUserById(id int32) *models.User
	AddUser(user *models.User)
	GetAllUsersData() []models.UserData
	RemoveUser(id int32)
}

type UserStorage struct {
	Users *sync.Map
}

func (st *UserStorage) AddUser(user *models.User) {
	st.Users.Store(user.UID, user)
}

func (st *UserStorage) RemoveUser(id int32) {
	st.Users.Delete(id)
}

func (st *UserStorage) GetUserById(id int32) *models.User {
	us, ok := st.Users.Load(id)
	if !ok {
		return nil
	}
	return us.(*models.User)
}

func (st *UserStorage) GetAllUsersData() []models.UserData {
	var usersData []models.UserData

	st.Users.Range(func(key, value interface{}) bool {
		if user, ok := value.(*models.User); ok {
			usersData = append(usersData, user.UserData)
		}
		return true
	})

	return usersData
}
