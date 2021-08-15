package repo_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/vbogretsov/guard/model"
	"github.com/vbogretsov/guard/repo"
)

var users = []model.User{
	{
		ID:      "123",
		Name:    "u0@mail.org",
		Created: 1000000000,
	},
	{
		ID:      "456",
		Name:    "u1@mail.org",
		Created: 1000000000,
	},
}

var refreshTokens = []model.RefreshToken{
	{
		ID:      "abc123",
		UserID:  "123",
		User:    users[0],
		Created: 1000000000,
		Expires: 1000000010,
	},
	{
		ID:      "abc456",
		UserID:  "456",
		User:    users[1],
		Created: 1000000000,
		Expires: 1000000010,
	},
}

var sessions = []model.Session{
	{
		ID:      "123",
		Value:   "session.123",
		Created: 1000000000,
		Expires: 1000000010,
	},
	{
		ID:      "456",
		Value:   "session.123",
		Created: 1000000010,
		Expires: 1000000020,
	},
}

func TestOnSQLite(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	require.NoError(t, err, "unable to create SQLite database")
	require.NoError(t, db.AutoMigrate(&model.User{}), "failed to auto migrate users")
	require.NoError(t, db.AutoMigrate(&model.RefreshToken{}), "failed to auto migrate refresh_tokens")
	require.NoError(t, db.AutoMigrate(&model.Session{}), "failed to auto migrate sessions")

	t.Run("Users", func(t *testing.T) {
		tx := repo.NewTransaction(db)
		ur := repo.NewUsers(tx)

		for _, u := range users {
			t.Run("Create", func(t *testing.T) {
				require.NoError(t, ur.Create(u), "failed to create user(%s)", u.ID)
			})
		}

		for _, u1 := range users {
			t.Run("Find", func(t *testing.T) {
				u2, err := ur.Find(u1.Name)
				require.NoError(t, err, "failed to find user(%s)", u1.ID)
				require.Equal(t, u1, u2, "the user found does not match expected one")
			})
		}

		t.Run("NotFind", func(t *testing.T) {
			u, err := ur.Find("xxx")
			require.Error(t, err, "found user(%s)", u.Name)
			require.ErrorIs(t, err, repo.ErrorNotFound)
		})
	})

	t.Run("RefreshTokens", func(t *testing.T) {
		tx := repo.NewTransaction(db)
		rr := repo.NewRefreshTokens(tx)

		for _, rt := range refreshTokens {
			t.Run("Create", func(t *testing.T) {
				require.NoError(t, rr.Create(rt), "failed to create refreshToken(%s)", rt.ID)
			})
		}

		for i, rt1 := range refreshTokens {
			t.Run("Find", func(t *testing.T) {
				rt2, err := rr.Find(rt1.ID)
				require.NoError(t, err, "failed to find refreshToken(%s)", rt1.ID)
				require.Equal(t, rt1, rt2, "the refreshToken found does not match expected one")
				require.Equal(t, users[i], rt2.User)
			})
		}

		t.Run("NotFind", func(t *testing.T) {
			rt, err := rr.Find("xxx")
			require.Error(t, err, "found refreshToken(%s)", rt.ID)
			require.ErrorIs(t, err, repo.ErrorNotFound)
		})

		t.Run("Delete", func(t *testing.T) {
			id0 := refreshTokens[0].ID
			require.NoError(t, rr.Delete(id0), "failed to delete refreshToken(%s)", id0)

			_, err := rr.Find(id0)
			require.Error(t, err)
			require.ErrorIs(t, err, repo.ErrorNotFound)

			id1 := refreshTokens[1].ID
			_, err = rr.Find(id1)
			require.NoError(t, err)
		})
	})

	t.Run("Sessions", func(t *testing.T) {
		tx := repo.NewTransaction(db)
		sr := repo.NewSessions(tx)

		for _, s := range sessions {
			t.Run("Create", func(t *testing.T) {
				require.NoError(t, sr.Create(s), "failed to create session(%s)", s.ID)
			})
		}

		for _, s1 := range sessions {
			t.Run("Find", func(t *testing.T) {
				s2, err := sr.Find(s1.ID)
				require.NoError(t, err, "failed to find session(%s)", s1.ID)
				require.Equal(t, s1, s2, "the xsrfToken found does not match expected one")
			})
		}

		t.Run("NotFind", func(t *testing.T) {
			xt, err := sr.Find("xxx")
			require.Error(t, err, "found xsrfToken(%s)", xt.ID)
			require.ErrorIs(t, err, repo.ErrorNotFound)
		})

		t.Run("Delete", func(t *testing.T) {
			id0 := sessions[0].ID
			require.NoError(t, sr.Delete(id0), "failed to delete session(%s)", id0)

			_, err := sr.Find(id0)
			require.Error(t, err)
			require.ErrorIs(t, err, repo.ErrorNotFound)

			id1 := sessions[1].ID
			_, err = sr.Find(id1)
			require.NoError(t, err)
		})
	})

	t.Run("Atomic", func(t *testing.T) {
		sess := model.Session{
			ID:      "atomic.session.123",
			Value:   "atomic.session.value.123",
			Created: 1600000000,
			Expires: 1600000010,
		}

		user := model.User{
			ID:      "atomic.user.123",
			Name:    "x0@mail.org",
			Created: 1600000000,
		}

		refresh := model.RefreshToken{
			ID:      "atomic.refresh.123",
			UserID:  user.ID,
			Created: 1600000000,
			Expires: 1600000010,
		}

		t.Run("Rollback", func(t *testing.T) {
			tx := repo.NewTransaction(db)

			sessRepo := repo.NewSessions(tx)
			userRepo := repo.NewUsers(tx)
			refreshRepo := repo.NewRefreshTokens(tx)

			test := func(tx repo.Transaction) {
				require.NoError(t, tx.Begin())
				defer func() { require.NoError(t, tx.Close()) }()

				require.NoError(t, sessRepo.Create(sess))
				require.NoError(t, userRepo.Create(user))
				require.NoError(t, refreshRepo.Create(refresh))
			}

			test(tx)

			_, err = sessRepo.Find(sess.ID)
			require.ErrorIs(t, err, repo.ErrorNotFound)

			_, err = userRepo.Find(user.Name)
			require.ErrorIs(t, err, repo.ErrorNotFound)

			_, err = refreshRepo.Find(refresh.ID)
			require.ErrorIs(t, err, repo.ErrorNotFound)

		})

		t.Run("Commit", func(t *testing.T) {
			tx := repo.NewTransaction(db)

			xsrfRepo := repo.NewSessions(tx)
			userRepo := repo.NewUsers(tx)
			refreshRepo := repo.NewRefreshTokens(tx)

			test := func(tx repo.Transaction) {
				require.NoError(t, tx.Begin())
				defer func() { require.NoError(t, tx.Close()) }()

				require.NoError(t, xsrfRepo.Create(sess))
				require.NoError(t, userRepo.Create(user))
				require.NoError(t, refreshRepo.Create(refresh))

				require.NoError(t, tx.Commit())
			}

			test(tx)

			_, err = xsrfRepo.Find(sess.ID)
			require.NoError(t, err)

			_, err = userRepo.Find(user.Name)
			require.NoError(t, err)

			_, err = refreshRepo.Find(refresh.ID)
			require.NoError(t, err)
		})
	})
}
