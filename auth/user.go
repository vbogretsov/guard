package auth

import (
	"errors"

	"github.com/vbogretsov/guard/model"
	"github.com/vbogretsov/guard/repo"
)

const uidLen = 64

type UserProvideCommand interface {
	Execute(username string) (model.User, error)
}

type userProvideCommand struct {
	users repo.Users
	timer Timer
}

func NewUserProvideCommand(users repo.Users, timer Timer) UserProvideCommand {
	return &userProvideCommand{users: users, timer: timer}
}

func (c *userProvideCommand) Execute(username string) (model.User, error) {
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
