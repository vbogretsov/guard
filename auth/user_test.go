package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/markbates/goth"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vbogretsov/guard/auth"
	"github.com/vbogretsov/guard/model"
	"github.com/vbogretsov/guard/repo"
)

type usersMock struct {
	mock.Mock
}

func (m *usersMock) Create(user model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *usersMock) Find(name string) (model.User, error) {
	args := m.Called(name)

	user := args.Get(0)
	if user == nil {
		return model.User{}, args.Error(1)
	}

	return user.(model.User), args.Error(1)
}

func matchUser(user model.User) func(model.User) bool {
	return func(arg model.User) bool {
		return user.Name == arg.Name &&
			user.Created == arg.Created
	}
}

type userFindOrCreatorMock struct {
	mock.Mock
}

func (m *userFindOrCreatorMock) FindOrCreate(username string) (model.User, error) {
	args := m.Called(username)

	value := args.Get(0)
	if value == nil {
		return model.User{}, args.Error(1)
	}

	return value.(model.User), args.Error(1)
}

type userFetcherMock struct {
	mock.Mock
}

func (m *userFetcherMock) Fetch(rawsess string, params goth.Params) (model.User, error) {
	args := m.Called(rawsess, params)

	user := args.Get(0)
	if user == nil {
		return model.User{}, args.Error(1)
	}

	return user.(model.User), args.Error(1)
}

type updaterMock struct {
	mock.Mock
}

func (m *updaterMock) Update(userID string, data map[string]interface{}) error {
	return m.Called(userID, data).Error(0)
}

func TestUserFinOrCreator(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		um := &usersMock{}
		tm := &timerMock{value: time.Now()}

		user := model.User{
			Name:    "u0@mail.org",
			Created: tm.Now().Unix(),
		}

		um.On("Find", user.Name).
			Return(model.User{}, repo.ErrorNotFound)
		um.On("Create", mock.MatchedBy(matchUser(user))).
			Return(nil)

		svc := auth.NewUserFindOrCreator(um, tm)

		result, err := svc.FindOrCreate(user.Name)
		require.NoError(t, err)
		require.NotEmpty(t, result.ID)
	})

	t.Run("Old", func(t *testing.T) {
		um := &usersMock{}
		tm := &timerMock{value: time.Now()}

		user := model.User{
			Name:    "u0@mail.org",
			Created: tm.Now().Unix(),
		}

		um.On("Find", user.Name).Return(user, nil)

		svc := auth.NewUserFindOrCreator(um, tm)

		result, err := svc.FindOrCreate(user.Name)
		require.NoError(t, err)
		require.Equal(t, user, result)
	})

	t.Run("FailOnFind", func(t *testing.T) {
		um := &usersMock{}
		tm := &timerMock{value: time.Now()}

		username := "xxx"
		fail := errors.New("xxx")

		um.On("Find", username).Return(model.User{}, fail)

		svc := auth.NewUserFindOrCreator(um, tm)
		_, err := svc.FindOrCreate(username)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("FailOnCreate", func(t *testing.T) {
		um := &usersMock{}
		tm := &timerMock{value: time.Now()}

		username := "xxx"
		fail := errors.New("xxx")

		um.On("Find", username).Return(model.User{}, repo.ErrorNotFound)
		um.On("Create", mock.Anything).Return(fail)

		svc := auth.NewUserFindOrCreator(um, tm)

		_, err := svc.FindOrCreate(username)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})
}

func TestUserFetcher(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		var params goth.Params

		provider := &providerMock{}
		session := &sessionMock{}
		userFoC := &userFindOrCreatorMock{}
		updater := &updaterMock{}

		rawsess := "user.session.value.123"

		gUser := goth.User{
			Email:   "u1@mail.org",
			RawData: map[string]interface{}{"Email": "u1@mail.org"},
		}

		user := model.User{
			ID:      "user.user.id",
			Name:    gUser.Email,
			Created: 1600000000,
		}

		provider.On("UnmarshalSession", rawsess).Return(session, nil)
		session.On("Authorize", provider, params).Return(nil, nil)
		provider.On("FetchUser", session).Return(gUser, nil)
		userFoC.On("FindOrCreate", gUser.Email).Return(user, nil)
		updater.On("Update", user.ID, gUser.RawData).Return(nil)

		cmd := auth.NewUserFetcher(provider, userFoC, updater)

		result, err := cmd.Fetch(rawsess, params)
		require.NoError(t, err)
		require.Equal(t, user, result)
	})

	t.Run("FailOnUnmarshal", func(t *testing.T) {
		var params goth.Params

		provider := &providerMock{}
		userFoC := &userFindOrCreatorMock{}

		rawsess := "user.session.value.123"
		fail := errors.New("xxx")

		provider.On("UnmarshalSession", rawsess).Return(nil, fail)

		cmd := auth.NewUserFetcher(provider, userFoC, &updaterMock{})

		_, err := cmd.Fetch(rawsess, params)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("FailOnAuthorize", func(t *testing.T) {
		var params goth.Params

		provider := &providerMock{}
		session := &sessionMock{}
		userFoC := &userFindOrCreatorMock{}

		rawsess := "user.session.value.123"

		provider.On("UnmarshalSession", rawsess).Return(session, nil)

		fail := errors.New("xxx")
		session.On("Authorize", provider, params).Return(nil, fail)

		cmd := auth.NewUserFetcher(provider, userFoC, &updaterMock{})

		_, err := cmd.Fetch(rawsess, params)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("FailOnFetch", func(t *testing.T) {
		var params goth.Params

		provider := &providerMock{}
		session := &sessionMock{}
		userFoC := &userFindOrCreatorMock{}

		rawsess := "user.session.value.123"
		fail := errors.New("xxx")

		provider.On("UnmarshalSession", rawsess).Return(session, nil)
		session.On("Authorize", provider, params).Return(nil, nil)
		provider.On("FetchUser", session).Return(nil, fail)

		cmd := auth.NewUserFetcher(provider, userFoC, &updaterMock{})

		_, err := cmd.Fetch(rawsess, params)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("FailOnFindOrCreate", func(t *testing.T) {
		var params goth.Params

		provider := &providerMock{}
		session := &sessionMock{}
		userFoC := &userFindOrCreatorMock{}

		rawsess := "user.session.value.123"
		fail := errors.New("xxx")

		gUser := goth.User{
			Email: "u1@mail.org",
		}

		provider.On("UnmarshalSession", rawsess).Return(session, nil)
		session.On("Authorize", provider, params).Return(nil, nil)
		provider.On("FetchUser", session).Return(gUser, nil)
		userFoC.On("FindOrCreate", gUser.Email).Return(nil, fail)

		cmd := auth.NewUserFetcher(provider, userFoC, &updaterMock{})

		_, err := cmd.Fetch(rawsess, params)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("FailOnUpdateProfile", func(t *testing.T) {
		var params goth.Params

		provider := &providerMock{}
		session := &sessionMock{}
		userFoC := &userFindOrCreatorMock{}
		updater := &updaterMock{}

		rawsess := "user.session.value.123"
		fail := errors.New("xxx")

		gUser := goth.User{
			Email:   "u1@mail.org",
			RawData: map[string]interface{}{"Email": "u1@mail.org"},
		}

		user := model.User{
			ID:      "user.user.id",
			Name:    gUser.Email,
			Created: 1600000000,
		}

		provider.On("UnmarshalSession", rawsess).Return(session, nil)
		session.On("Authorize", provider, params).Return(nil, nil)
		provider.On("FetchUser", session).Return(gUser, nil)
		userFoC.On("FindOrCreate", gUser.Email).Return(user, nil)
		updater.On("Update", user.ID, gUser.RawData).Return(fail)

		cmd := auth.NewUserFetcher(provider, userFoC, updater)

		_, err := cmd.Fetch(rawsess, params)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})
}
