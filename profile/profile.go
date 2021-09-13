package profile

type Updater interface {
	Update(userID string, data map[string]interface{}) error
}

type empty struct{}

func Empty() Updater {
	return &empty{}
}

func (e *empty) Update(userID string, data map[string]interface{}) error {
	return nil
}

type Claimer interface {
	GetClaims(userID string) (map[string]interface{}, error)
}
