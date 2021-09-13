package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rs/zerolog"

	"github.com/vbogretsov/guard"
)

type ErrorLogMarshaler guard.Error

func (err ErrorLogMarshaler) MarshalZerologObject(e *zerolog.Event) {
	e.Str(zerolog.ErrorFieldName, err.Err.Error())
	raw, ex := json.Marshal(err.Ctx)
	if ex != nil {
		e.RawJSON("ctx", raw)
		return
	}
	e.Str("cxt", fmt.Sprintf("failed to serialize error: %v", ex))
}

func ErrorMarshalFunc(err error) interface{} {
	e := guard.Error{}
	if errors.As(err, &e) {
		return ErrorLogMarshaler(e)
	}
	return err
}
