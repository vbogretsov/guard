package repo

import (
	"gorm.io/gorm"

	"github.com/vbogretsov/guard/model"
)

var ErrorNotFound = gorm.ErrRecordNotFound

type Users interface {
	Find(name string) (model.User, error)
	Create(user model.User) error
}

type RefreshTokens interface {
	Find(value string) (model.RefreshToken, error)
	Create(token model.RefreshToken) error
	Delete(value string) error
}

type XSRFTokens interface {
	Find(value string) (model.XSRFToken, error)
	Create(token model.XSRFToken) error
	Delete(value string) error
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

type xsrfTokens struct {
	db *gorm.DB
}

func NewXSRFTokens(db *gorm.DB) XSRFTokens {
	return &xsrfTokens{db: db}
}

func (x *xsrfTokens) Find(value string) (model.XSRFToken, error) {
	var token model.XSRFToken

	r := x.db.First(&token, "id = ?", value)
	if r.Error != nil {
		return token, r.Error
	}

	return token, nil
}

func (x *xsrfTokens) Create(token model.XSRFToken) error {
	return x.db.Create(&token).Error
}

func (x *xsrfTokens) Delete(value string) error {
	token := model.XSRFToken{ID: value}
	return x.db.Delete(&token).Error
}
