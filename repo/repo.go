package repo

import (
	"gorm.io/gorm"

	"github.com/vbogretsov/guard/model"
)

var ErrorNotFound = gorm.ErrRecordNotFound

type Transaction interface {
	Begin() error
	Commit() error
	Close() error
}

type Users interface {
	Find(name string) (model.User, error)
	Create(user model.User) error
}

type RefreshTokens interface {
	Find(value string) (model.RefreshToken, error)
	Create(token model.RefreshToken) error
	Delete(value string) error
}

type Sessions interface {
	Find(code string) (model.Session, error)
	Create(sess model.Session) error
	Delete(code string) error
}

type users struct {
	db *gorm.DB
}

func NewUsers(db *gorm.DB) Users {
	return &users{db: db}
}

func (u *users) Create(user model.User) error {
	return u.db.Create(&user).Error
}

func (u *users) Find(name string) (model.User, error) {
	var user model.User

	r := u.db.First(&user, "name = ?", name)
	if r.Error != nil {
		return user, r.Error
	}

	return user, nil
}

type refreshTokens struct {
	db *gorm.DB
}

func NewRefreshTokens(db *gorm.DB) RefreshTokens {
	return &refreshTokens{db: db}
}

func (rt *refreshTokens) Create(token model.RefreshToken) error {
	return rt.db.Create(&token).Error
}

func (rt *refreshTokens) Find(id string) (model.RefreshToken, error) {
	var token model.RefreshToken

	r := rt.db.Joins("User").First(&token, "refresh_tokens.id = ?", id)
	if r.Error != nil {
		return token, r.Error
	}

	return token, nil
}

func (rt *refreshTokens) Delete(id string) error {
	token := model.RefreshToken{ID: id}
	return rt.db.Delete(&token).Error
}

type sessions struct {
	db *gorm.DB
}

func NewSessions(db *gorm.DB) Sessions {
	return &sessions{db: db}
}

func (s *sessions) Find(value string) (model.Session, error) {
	var sess model.Session

	r := s.db.First(&sess, "id = ?", value)
	if r.Error != nil {
		return sess, r.Error
	}

	return sess, nil
}

func (s *sessions) Create(sess model.Session) error {
	return s.db.Create(&sess).Error
}

func (s *sessions) Delete(code string) error {
	sess := model.Session{ID: code}
	return s.db.Delete(&sess).Error
}
