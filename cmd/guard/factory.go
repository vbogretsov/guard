package main

import (
	"time"

	"github.com/markbates/goth"
	"gorm.io/gorm"

	"github.com/vbogretsov/guard/auth"
	"github.com/vbogretsov/guard/repo"
)

type FactoryConfig struct {
	SecretKey  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	CodeTTL    time.Duration
}

type factory struct {
	db  *gorm.DB
	cfg FactoryConfig
}

type scope struct {
	db       *gorm.DB
	tx       *repo.GormTx
	cfg      FactoryConfig
	timer    auth.Timer
	users    repo.Users
	tokens   repo.RefreshTokens
	sessions repo.Sessions
}

func NewFactory(db *gorm.DB, cfg FactoryConfig) auth.Factory {
	return &factory{
		db:  db,
		cfg: cfg,
	}
}

func (f *factory) NewOAuthStarter(provider goth.Provider) auth.OAuthStarter {
	return f.scope().newOAuthStarter(provider)
}

func (f *factory) NewSignIner(provider goth.Provider) auth.SignIner {
	return f.scope().newSignIner(provider)
}

func (f *factory) NewRefresher() auth.Refresher {
	return f.scope().newRefresher()
}

func (f *factory) scope() *scope {
	return &scope{db: f.db, cfg: f.cfg}
}

func (s *scope) newTimer() auth.Timer {
	if s.timer == nil {
		s.timer = &auth.RealTimer{}
	}
	return s.timer
}

func (s *scope) newTransaction() *repo.GormTx {
	if s.tx == nil {
		s.tx = repo.NewTransaction(s.db)
	}
	return s.tx
}

func (s *scope) newUsersRepo() repo.Users {
	if s.users == nil {
		s.users = repo.NewUsers(s.newTransaction())
	}
	return s.users
}

func (s *scope) newRefreshTokensRepo() repo.RefreshTokens {
	if s.tokens == nil {
		s.tokens = repo.NewRefreshTokens(s.newTransaction())
	}
	return s.tokens
}

func (s *scope) newSessionsRepo() repo.Sessions {
	if s.sessions == nil {
		s.sessions = repo.NewSessions(s.newTransaction())
	}
	return s.sessions
}

func (s *scope) newSessionValidator() auth.SessionValidator {
	return auth.NewSessionValidator(
		s.newSessionsRepo(),
		s.newTimer(),
	)
}

func (s *scope) newUserFindOrCreator() auth.UserFindOrCreator {
	return auth.NewUserFindOrCreator(
		s.newUsersRepo(),
		s.newTimer(),
	)
}

func (s *scope) newRefreshTokensCreator() auth.RefreshTokenCreator {
	return auth.NewRefreshTokenCreator(
		s.newRefreshTokensRepo(),
		s.newTimer(),
		s.cfg.RefreshTTL,
	)
}

func (s *scope) newIssuer() auth.Issuer {
	return auth.NewIssuer(
		s.cfg.SecretKey,
		s.newTimer(),
		s.cfg.AccessTTL,
		s.newRefreshTokensCreator(),
	)
}

func (s *scope) newRefresher() auth.Refresher {
	return auth.NewRefresher(
		s.newTimer(),
		s.newRefreshTokensRepo(),
		s.newIssuer(),
	)
}

func (s *scope) newUserFetcher(provider goth.Provider) auth.UserFetcher {
	return auth.NewUserFetcher(
		provider,
		s.newUserFindOrCreator(),
	)
}

func (s *scope) newSignIner(provider goth.Provider) auth.SignIner {
	return auth.NewSignIner(
		s.newSessionValidator(),
		s.newUserFetcher(provider),
		s.newIssuer(),
	)
}

func (s *scope) newOAuthStarter(provider goth.Provider) auth.OAuthStarter {
	return auth.NewOAuthStarter(
		s.cfg.CodeTTL,
		s.newTimer(),
		s.newSessionsRepo(),
		provider,
	)
}
