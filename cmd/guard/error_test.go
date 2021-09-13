package main

import (
	// "encoding/json"
	// "errors"
	"testing"

	"github.com/rs/zerolog"
	// "github.com/stretchr/testify/require"

	// "github.com/vbogretsov/guard"
)

type levelWriter struct {
	buf []byte
}

func (lv *levelWriter) Write(p []byte) (n int, err error) {
	lv.buf = p
	n = len(p)
	err = nil
	return
}

func (lv *levelWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	return lv.Write(p)
}

func TestErrorLogMarshaler(t *testing.T) {
	// zerolog.ErrorMarshalFunc = ErrorMarshalFunc

	/* t.Run("GuardError", func(t *testing.T) {
		w := &levelWriter{}
		l := zerolog.New(w)

		err := guard.Error{
			Err: errors.New("description"),
			Ctx: map[string]interface{}{
				"user_id": "123",
			},
		}

		l.Err(err).Send()

		v := map[string]interface{}{}
		require.NoError(t, json.Unmarshal(w.buf, &v))
		require.Equal(t, err.Ctx, v["ctx"])
	})

	t.Run("PlainError", func(t *testing.T) {
		w := &levelWriter{}
		l := zerolog.New(w)

		err := errors.New("xxx")

		l.Err(err).Send()
		
		v := map[string]interface{}{}
		require.NoError(t, json.Unmarshal(w.buf, &v))
		require.Equal(t, err.Error(), v["err"])
	}) */
}

