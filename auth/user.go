package auth

import (
	"errors"

	"github.com/vbogretsov/guard/model"
	"github.com/vbogretsov/guard/repo"
)

const uidLen = 64

type UserFindOrCreator interface {
	FindOrCreate(username string) (model.User, error)
}

type userFindOrCreator struct {
	users repo.Users
	timer Timer
}

func NewFindOrCreator(users repo.Users, timer Timer) UserFindOrCreator {
	return &userFindOrCreator{users: users, timer: timer}
}

func (c *userFindOrCreator) FindOrCreate(username string) (model.User, error) {
	user, err := c.users.Find(username)
	if err != nil {
		if !errors.Is(err, repo.ErrorNotFound) {
			return user, err
		}

		uid, err := generateRandomString(uidLen)
		if err != nil {
			return user, err
		}

		user.ID = uid
		user.Name = username
		user.Created = c.timer.Now().Unix()

		if err := c.users.Create(user); err != nil {
			return user, err
		}
	}
	return user, nil
}
