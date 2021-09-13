package guard

type Error struct {
	Err error
	Ctx map[string]interface{}
}

func (e Error) Error() string {
	return e.Err.Error()
}
