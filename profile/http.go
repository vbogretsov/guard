package profile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/PaesslerAG/jsonpath"
	"github.com/rs/zerolog/log"
)

type httpClaimer struct {
	cli     http.Client
	url     string
	jspath  string
	authHdr string
	authVal string
}

func NewHttpClaimer(url, jspath, authHdr, authVal string) Claimer {
	return &httpClaimer{
		cli:     *http.DefaultClient,
		url:     url,
		jspath:  jspath,
		authHdr: authHdr,
		authVal: authVal,
	}
}

func (c *httpClaimer) GetClaims(userID string) (map[string]interface{}, error) {
	log := log.With().Str("userID", userID).Logger()
	req, err := http.NewRequest(http.MethodGet, c.url+userID, nil)
	if err != nil {
		log.Err(err).Msg("failed to create claims request")
		return nil, fmt.Errorf("failed to create claims request: %w", err)
	}

	req.Header.Add(c.authHdr, c.authVal)

	res, err := c.cli.Do(req)
	if err != nil {
		log.Err(err).Msg("failed to get claims")
		return nil, fmt.Errorf("failed to get claims: %w", err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Err(err).Msg("failed to read claims response")
		return nil, fmt.Errorf("failed to read claims response: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		log.Err(err).Int("status_code", res.StatusCode).Msg("claims request failed")
		return nil, fmt.Errorf("claims request failed: %s", string(body))
	}

	var value map[string]interface{}
	if err := json.Unmarshal(body, &value); err != nil {
		log.Err(err).Msg("failed to parse claims response")
		return nil, fmt.Errorf("failed to parse claims response: %w", err)
	}

	claims, err := jsonpath.Get(c.jspath, value)
	if err != nil {
		log.Err(err).Msg("failed to find claims")
		return nil, fmt.Errorf("failed to find claims: %w", err)
	}

	obj, ok := claims.(map[string]interface{})
	if !ok {
		log.Err(err).Msg("failed to read claims: expected object")
		return nil, fmt.Errorf("failed to read claims: expected object")
	}

	return obj, nil
}
