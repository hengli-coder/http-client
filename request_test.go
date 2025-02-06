package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetWaitWithExponentialBackoff(t *testing.T) {
	client := Retry(WithAttempt(5), WithWait(1*time.Second), WithMaxWait(8*time.Second))

	for i := 0; i < 5; i++ {
		client.curAttempt = i
		waitTime := client.getWait()

		if waitTime > client.maxWait {
			t.Fatalf("wait time exceeded max wait limit: got %v, expected less than or equal to %v", waitTime, client.maxWait)
		}
	}
}

func TestClientDoWithContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Simulate a long running request
		time.Sleep(3 * time.Second)
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	client := Retry()
	request := NewReq(http.MethodGet, server.URL).SetContext(ctx)

	_, err := client.Do(request)
	if err == nil {
		t.Fatal("expected an error due to context cancellation")
	}
}

type TestStruct struct {
	HeaderField string `header:"X-Test-Header"`
	QueryField  string `query:"query_param"`
}

func TestRequestEncoding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("X-Test-Header") != "HeaderValue" {
			t.Error("Header encoding failed")
		}
		if req.URL.Query().Get("query_param") != "QueryParamValue" {
			t.Error("Query param encoding failed")
		}
		rw.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := Retry()
	request := NewReq(http.MethodGet, server.URL).
		SetHeader(&TestStruct{HeaderField: "HeaderValue"}).
		SetQuery(&TestStruct{QueryField: "QueryParamValue"})

	_, _ = client.Do(request)
}

func TestResponseDecoding(t *testing.T) {
	respBody := `{"name": "test", "age": 25}`
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write([]byte(respBody))
	}))
	defer server.Close()

	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	client := Retry()
	request := NewReq(http.MethodGet, server.URL).BindBody(&Person{})

	_, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}

	person, ok := request.result.(*Person)
	if !ok {
		t.Errorf("the result is not *Person type")
	}

	if person.Name != "test" || person.Age != 25 {
		t.Fatalf("Decoding failed: expected {Name: 'test', Age: 25}, got %+v", *person)
	}
}
