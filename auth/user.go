package auth

import (
	"errors"

	"github.com/markbates/goth"

	"github.com/vbogretsov/guard/model"
	"github.com/vbogretsov/guard/repo"
)

const uidLen = 64

type UserFetcher interface {
	Fetch(session string, params goth.Params) (model.User, error)
}

type userFetcher struct {
	provider goth.Provider
	users    UserFindOrCreator
}

func NewUserFetcher(provider goth.Provider, users UserFindOrCreator) UserFetcher {
	return &userFetcher{
		provider: provider,
		users:    users,
	}
}

func (c *userFetcher) Fetch(rawsess string, params goth.Params) (model.User, error) {
	var empty model.User

	session, err := c.provider.UnmarshalSession(rawsess)
	if err != nil {
		return empty, err
	}

	_, err = session.Authorize(c.provider, params)
	if err != nil {
		return empty, err
	}

	gUser, err := c.provider.FetchUser(session)
	if err != nil {
		return empty, err
	}

	user, err := c.users.FindOrCreate(gUser.Email)
	if err != nil {
		return empty, err
	}

	// TODO: POST profileService/users

	return user, err
}

type UserFindOrCreator interface {
	FindOrCreate(username string) (model.User, error)
}

type userFindOrCreator struct {
	users repo.Users
	timer Timer
}

func NewUserFindOrCreator(users repo.Users, timer Timer) UserFindOrCreator {
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
