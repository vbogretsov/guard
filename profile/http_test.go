package profile_test

import (
	"testing"

	"gopkg.in/h2non/gock.v1"

	"github.com/stretchr/testify/require"
	"github.com/vbogretsov/guard/profile"
)

func TestHttpClaimer(t *testing.T) {
	url := "http://guard.example.com"
	endpoint := "/claims"
	jspath := "$.user_by_pk"
	authHdr := "Authorization"
	authKey := "^abcde.abcde$"
	userID := "123"

	t.Run("Success", func(t *testing.T) {
		defer gock.Clean()

		claims := `{
			"user_by_pk": {
				"claims": {
					"role": "admin",
					"x-hasura-allowed-roles": ["admin"],
					"x-hasura-default-role": "admin",
					"x-hasura-user-id": "123"
				}
			}
		}`

		gock.
			New(url).
			Get(endpoint).
			MatchParams(map[string]string{"user_id": userID}).
			MatchHeader(authHdr, authKey).
			Reply(200).
			BodyString(claims)

		claimer := profile.NewHttpClaimer(url+endpoint+"?user_id=", jspath, authHdr, authKey)

		resp, err := claimer.GetClaims(userID)
		require.NoError(t, err)

		require.Equal(t, resp, map[string]interface{}{
			"claims": map[string]interface{}{
				"role":                   "admin",
				"x-hasura-allowed-roles": []interface{}{"admin"},
				"x-hasura-default-role":  "admin",
				"x-hasura-user-id":       userID,
			},
		})
	})

	t.Run("FailedRequest", func(t *testing.T) {
		defer gock.Clean()

		gock.
			New(url).
			Get(endpoint).
			MatchHeader(authHdr, authKey).
			Reply(200)

		claimer := profile.NewHttpClaimer(url+endpoint+"?user_id=", jspath, authHdr, authKey)

		_, err := claimer.GetClaims(userID)
		require.Error(t, err)
	})

	t.Run("BadRequest", func(t *testing.T) {
		defer gock.Clean()

		gock.
			New(url).
			Get(endpoint).
			MatchParams(map[string]string{"user_id": userID}).
			MatchHeader(authHdr, authKey).
			Reply(400).
			BodyString(`{"message": "user not found"}`)

		claimer := profile.NewHttpClaimer(url+endpoint+"?user_id=", jspath, authHdr, authKey)

		_, err := claimer.GetClaims(userID)
		require.Error(t, err)
	})

	t.Run("FailedToParseBody", func(t *testing.T) {
		defer gock.Clean()

		claims := `{
			"user_by_pk": {
				"claims": {
					"role": "admin",
					"x-hasura-allowed-roles": ["admin"],
					"x-hasura-default-role": "admin",
					"x-hasura-user-id": "123",
				}
			}
		}`

		gock.
			New(url).
			Get(endpoint).
			MatchParams(map[string]string{"user_id": userID}).
			MatchHeader(authHdr, authKey).
			Reply(200).
			BodyString(claims)

		claimer := profile.NewHttpClaimer(url+endpoint+"?user_id=", jspath, authHdr, authKey)

		_, err := claimer.GetClaims(userID)
		require.Error(t, err)
	})

	t.Run("FailedJsonPath", func(t *testing.T) {
		defer gock.Clean()

		claims := `{
			"user": {
				"claims": {
					"role": "admin",
					"x-hasura-allowed-roles": ["admin"],
					"x-hasura-default-role": "admin",
					"x-hasura-user-id": "123"
				}
			}
		}`

		gock.
			New(url).
			Get(endpoint).
			MatchParams(map[string]string{"user_id": userID}).
			MatchHeader(authHdr, authKey).
			Reply(200).
			BodyString(claims)

		claimer := profile.NewHttpClaimer(url+endpoint+"?user_id=", jspath, authHdr, authKey)

		_, err := claimer.GetClaims(userID)
		require.Error(t, err)
	})

	t.Run("ClaimsNotAnObject", func(t *testing.T) {
		defer gock.Clean()

		claims := `{
			"user_by_pk": [
				"admin",
				["admin"],
				"admin",
				"123"
			]
		}`

		gock.
			New(url).
			Get(endpoint).
			MatchParams(map[string]string{"user_id": userID}).
			MatchHeader(authHdr, authKey).
			Reply(200).
			BodyString(claims)

		claimer := profile.NewHttpClaimer(url+endpoint+"?user_id=", jspath, authHdr, authKey)

		_, err := claimer.GetClaims(userID)
		require.Error(t, err)
	})
}

func TestHttpUpdater(t *testing.T) {
	url := "http://guard.example.com"
	endpoint := "/users"
	authHdr := "Authorization"
	authKey := "^abcde.abcde$"
	userID := "123"

	t.Run("Success", func(t *testing.T) {
		defer gock.Clean()

		gock.
			New(url).
			Post(endpoint).
			MatchHeader(authHdr, authKey).
			Reply(200)

		updater := profile.NewHttpUpdater(url+endpoint, authHdr, authKey)

		err := updater.Update(userID, map[string]interface{}{
			"id": userID,
			"data": map[string]interface{}{
				"first_name": "Don",
				"last_name": "Knuth",
			},
		})
		require.NoError(t, err)
	})
}
