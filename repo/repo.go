package repo

import (
	"errors"

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

type XSRFTokens interface {
	Find(value string) (model.XSRFToken, error)
	Create(token model.XSRFToken) error
	Delete(value string) error
}

type GormTx struct {
	stack []*gorm.DB
	open  int
	clos  int
}

func NewTransaction(db *gorm.DB) *GormTx {
	return &GormTx{
		stack: []*gorm.DB{db},
		open:  0,
		clos:  0,
	}
}

func (tx *GormTx) db() *gorm.DB {
	return tx.stack[tx.open]
}

func (tx *GormTx) Begin() error {
	db := tx.db().Begin()
	if db.Error != nil {
		return db.Error
	}

	tx.stack = append(tx.stack, db)
	tx.open++
	tx.clos++

	return nil
}

func (tx *GormTx) Commit() error {
	if tx.open == 0 {
		return errors.New("commit failed because transactoin wasn't started")
	}

	id := tx.open
	db := tx.db().Commit()
	tx.open--

	if db.Error != nil {
		return db.Error
	}

	tx.stack[id] = nil
	return nil
}

func (tx *GormTx) Close() error {
	if tx.clos == 0 {
		return nil
	}

	if tx.stack[tx.clos] == nil {
		return nil
	}

	db := tx.stack[tx.clos].Rollback()
	tx.clos--
	tx.open = tx.clos

	return db.Error
}

type users struct {
	tx *GormTx
}

func NewUsers(tx *GormTx) Users {
	return &users{tx: tx}
}

func (u *users) Create(user model.User) error {
	return u.tx.db().Create(&user).Error
}

func (u *users) Find(name string) (model.User, error) {
	var user model.User

	r := u.tx.db().First(&user, "name = ?", name)
	if r.Error != nil {
		return user, r.Error
	}

	return user, nil
}

type refreshTokens struct {
	tx *GormTx
}

func NewRefreshTokens(tx *GormTx) RefreshTokens {
	return &refreshTokens{tx: tx}
}

func (rt *refreshTokens) Create(token model.RefreshToken) error {
	return rt.tx.db().Create(&token).Error
}

func (rt *refreshTokens) Find(id string) (model.RefreshToken, error) {
	var token model.RefreshToken

	r := rt.tx.db().Joins("User").First(&token, "refresh_tokens.id = ?", id)
	if r.Error != nil {
		return token, r.Error
	}

	return token, nil
}

func (rt *refreshTokens) Delete(id string) error {
	token := model.RefreshToken{ID: id}
	return rt.tx.db().Delete(&token).Error
}

type xsrfTokens struct {
	tx *GormTx
}

func NewXSRFTokens(tx *GormTx) XSRFTokens {
	return &xsrfTokens{tx: tx}
}

func (x *xsrfTokens) Find(value string) (model.XSRFToken, error) {
	var token model.XSRFToken

	r := x.tx.db().First(&token, "id = ?", value)
	if r.Error != nil {
		return token, r.Error
	}

	return token, nil
}

func (x *xsrfTokens) Create(token model.XSRFToken) error {
	return x.tx.db().Create(&token).Error
}

func (x *xsrfTokens) Delete(value string) error {
	token := model.XSRFToken{ID: value}
	return x.tx.db().Delete(&token).Error
}
