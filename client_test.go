package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_Do(t *testing.T) {
	s := mock()
	defer s.Close()
	var ctx = context.Background()
	r := NewReqTLS(http.MethodGet, s.URL).SetRetryStrategy(DefaultRetryStrategy).
		SetContext(ctx).SetBody([]byte(`{"test":true'}`))
	response, err := Retry(WithHttpClient(s.Client()), WithAttempt(3), WithMaxWait(time.Second*35)).Do(r)
	if err == nil {
		t.Errorf("TestClient_Do() error = %v, wantErr %v", err, true)
	}

	if response != nil {
		t.Errorf("TestClient_Do() response = %v, want nil %v", response, true)
	}
}

func TestClient_Do_timeout(t *testing.T) {
	s := mock()
	defer s.Close()

	var ctx = context.Background()
	r := NewReqTLS(http.MethodGet, s.URL).SetRetryStrategy(DefaultRetryStrategy).
		SetContext(ctx).SetBody([]byte(`{"test":true'}`))

	response, err := Retry(WithHttpClient(s.Client()), WithAttempt(4), WithMaxWait(time.Second*10)).Do(r)
	if err == nil {
		t.Errorf("TestClient_Do() error = %v, wantErr %v", err, true)
	}

	if response != nil {
		t.Errorf("TestClient_Do() response = %v, want nil %v", response, true)
	}
}

func mock() *httptest.Server {

	handle := http.NewServeMux()

	handle.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte("test"))
	})

	server := httptest.NewTLSServer(handle)

	return server
}

func Test_loadCACert(t *testing.T) {

}

func TestClient_getWait(t *testing.T) {
	tests := []struct {
		name       string
		curAttempt int
		wait       time.Duration
		maxWait    time.Duration
	}{
		{"Initial attempt", 1, time.Millisecond * 1000, time.Second * 30},
		{"Second attempt", 2, time.Millisecond * 1000, time.Second * 30},
		{"Exponential backoff", 3, time.Millisecond * 1000, time.Second * 30},
		{"Max wait exceeded", 10, time.Millisecond * 1000, time.Second * 30},
		// {"Zero max wait", 1, time.Millisecond * 1000, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				wait:       tt.wait,
				maxWait:    tt.maxWait,
				curAttempt: tt.curAttempt,
			}

			got := client.getWait()
			assert.LessOrEqual(t, got, tt.maxWait)
			assert.GreaterOrEqual(t, got, tt.wait/2)
		})
	}
}
