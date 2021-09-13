package profile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/PaesslerAG/jsonpath"

	"github.com/vbogretsov/guard"
)

type httpClaimer struct {
	cli      *http.Client
	endpoint string
	jspath   string
	authHdr  string
	authKey  string
}

func NewHttpClaimer(endpoint, jspath, authHdr, authKey string) Claimer {
	return &httpClaimer{
		cli:      http.DefaultClient,
		endpoint: endpoint,
		jspath:   jspath,
		authHdr:  authHdr,
		authKey:  authKey,
	}
}

func (c *httpClaimer) GetClaims(userID string) (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, c.endpoint+userID, nil)
	if err != nil {
		return nil, guard.Error{
			Err: fmt.Errorf("failed to create request: %w", err),
			Ctx: map[string]interface{}{
				"user_id": userID,
			},
		}
	}

	req.Header.Add(c.authHdr, c.authKey)

	res, err := c.cli.Do(req)
	if err != nil {
		return nil, guard.Error{
			Err: fmt.Errorf("failed to execute request: %w", err),
			Ctx: map[string]interface{}{
				"user_id": userID,
			},
		}
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, guard.Error{
			Err: fmt.Errorf("failed to read response: %w", err),
			Ctx: map[string]interface{}{
				"user_id": userID,
			},
		}
	}

	var value map[string]interface{}
	if err := json.Unmarshal(body, &value); err != nil {
		return nil, guard.Error{
			Err: fmt.Errorf("failed to parse response: %w", err),
			Ctx: map[string]interface{}{
				"user_id":     userID,
				"status_code": res.StatusCode,
				"body":        string(body),
			},
		}
	}

	if res.StatusCode != http.StatusOK {
		return nil, guard.Error{
			Err: errors.New("request failed"),
			Ctx: map[string]interface{}{
				"user_id":     userID,
				"status_code": res.StatusCode,
				"body":        value,
			},
		}
	}

	claims, err := jsonpath.Get(c.jspath, value)
	if err != nil {
		return nil, guard.Error{
			Err: fmt.Errorf("failed to find claims: %w", err),
			Ctx: map[string]interface{}{
				"user_id":     userID,
				"status_code": res.StatusCode,
				"body":        value,
			},
		}
	}

	obj, ok := claims.(map[string]interface{})
	if !ok {
		return nil, guard.Error{
			Err: fmt.Errorf("failed to read claims: expected object"),
			Ctx: map[string]interface{}{
				"user_id":     userID,
				"status_code": res.StatusCode,
				"body":        body,
			},
		}
	}

	return obj, nil
}

type httpUpdater struct {
	cli      *http.Client
	endpoint string
	authHdr  string
	authKey  string
}

func NewHttpUpdater(endpoint, authHdr, authKey string) Updater {
	return &httpUpdater{
		cli:      http.DefaultClient,
		endpoint: endpoint,
		authHdr:  authHdr,
		authKey:  authKey,
	}
}

func (c *httpUpdater) Update(userID string, data map[string]interface{}) error {
	buf, err := json.Marshal(map[string]interface{}{
		"id":   userID,
		"data": data,
	})
	if err != nil {
		return guard.Error{
			Err: fmt.Errorf("failed to serialize user data: %w", err),
			Ctx: map[string]interface{}{
				"user_id": userID,
			},
		}
	}

	req, err := http.NewRequest(http.MethodPost, c.endpoint, bytes.NewReader(buf))
	if err != nil {
		return guard.Error{
			Err: fmt.Errorf("failed to create request: %w", err),
			Ctx: map[string]interface{}{
				"user_id": userID,
				"data":    data,
			},
		}
	}

	req.Header.Add(c.authHdr, c.authKey)

	res, err := c.cli.Do(req)
	if err != nil {
		return guard.Error{
			Err: fmt.Errorf("failed to execute request: %w", err),
			Ctx: map[string]interface{}{
				"user_id": userID,
				"data":    data,
			},
		}
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return guard.Error{
				Err: fmt.Errorf("failed to read response: %w", err),
				Ctx: map[string]interface{}{
					"user_id":     userID,
					"data":        data,
					"status_code": res.StatusCode,
				},
			}
		}
		return guard.Error{
			Err: errors.New("failed to update profile"),
			Ctx: map[string]interface{}{
				"user_id":     userID,
				"data":        data,
				"status_code": res.StatusCode,
				"body":        body,
			},
		}
	}

	return nil
}
