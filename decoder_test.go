package client

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_decodeRawBody(t *testing.T) {
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader("raw body content")),
	}
	req := &Req{
		result: new(string),
	}

	err := decodeRawBody(resp, req)
	assert.NoError(t, err)
	assert.Equal(t, "raw body content", *(req.result.(*string)))
}

func Test_decodeJsonBody(t *testing.T) {
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader(`{"key": "value"}`)),
	}
	req := &Req{
		result: &map[string]string{},
	}

	err := decodeJsonBody(resp, req)
	assert.NoError(t, err)
	assert.Equal(t, map[string]string{"key": "value"}, *(req.result.(*map[string]string)))
}

func Test_decodeBody(t *testing.T) {
	t.Run("decode raw body", func(t *testing.T) {
		resp := &http.Response{
			Body: io.NopCloser(strings.NewReader("raw body content")),
		}
		req := &Req{
			result: new(string),
		}

		err := decodeBody(resp, req)
		assert.NoError(t, err)
		assert.Equal(t, "raw body content", *(req.result.(*string)))
	})

	t.Run("decode json body", func(t *testing.T) {
		resp := &http.Response{
			Body: io.NopCloser(strings.NewReader(`{"key": "value"}`)),
		}
		req := &Req{
			result: &map[string]string{},
		}

		err := decodeBody(resp, req)
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{"key": "value"}, *(req.result.(*map[string]string)))
	})
}
