package auth

import (
	"time"

	"github.com/markbates/goth"
	"gorm.io/gorm"

	"github.com/vbogretsov/guard/repo"
)

type Config struct {
	SecretKey  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	CodeTTL    time.Duration
}

type factory struct {
	db  *gorm.DB
	cfg Config
}

type scope struct {
	db       *gorm.DB
	tx       *repo.GormTx
	cfg      Config
	timer    Timer
	users    repo.Users
	tokens   repo.RefreshTokens
	sessions repo.Sessions
}

func NewFactory(db *gorm.DB, cfg Config) Factory {
	return &factory{
		db:  db,
		cfg: cfg,
	}
}

func (f *factory) NewOAuthStarter(provider goth.Provider) OAuthStarter {
	return f.scope().newOAuthStarter(provider)
}

func (f *factory) NewSignIner(provider goth.Provider) SignIner {
	return f.scope().newSignIner(provider)
}

func (f *factory) NewRefresher() Refresher {
	return f.scope().newRefresher()
}

func (f *factory) scope() *scope {
	return &scope{db: f.db, cfg: f.cfg}
}

func (s *scope) newTimer() Timer {
	if s.timer == nil {
		s.timer = &RealTimer{}
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

func (s *scope) newSessionValidator() SessionValidator {
	return NewSessionValidator(s.newSessionsRepo(), s.newTimer())
}

func (s *scope) newUserFindOrCreator() UserFindOrCreator {
	return NewUserFindOrCreator(s.newUsersRepo(), s.newTimer())
}

func (s *scope) newRefreshTokensCreator() RefreshTokenCreator {
	return NewRefreshTokenCreator(s.newRefreshTokensRepo(), s.newTimer(), s.cfg.RefreshTTL)
}

func (s *scope) newIssuer() Issuer {
	return NewIssuer(s.cfg.SecretKey, s.newTimer(), s.cfg.AccessTTL, s.newRefreshTokensCreator())
}

func (s *scope) newRefresher() Refresher {
	return NewRefresher(s.newTimer(), s.newRefreshTokensRepo(), s.newIssuer())
}

func (s *scope) newUserFetcher(provider goth.Provider) UserFetcher {
	return NewUserFetcher(provider, s.newUserFindOrCreator())
}

func (s *scope) newSignIner(provider goth.Provider) SignIner {
	return NewSignIner(s.newSessionValidator(), s.newUserFetcher(provider), s.newIssuer())
}

func (s *scope) newOAuthStarter(provider goth.Provider) OAuthStarter {
	return NewOAuthStarter(s.cfg.CodeTTL, s.newTimer(), s.newSessionsRepo(), provider)
}
