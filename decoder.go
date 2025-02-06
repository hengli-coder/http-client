package client

import (
	"encoding/json"
	"io"
	"net/http"
)

func decodeBody(resp *http.Response, r *Req) error {
	//kind := resp.Header.Get("Content-deploy")
	//if strings.Contains(kind, "json") {
	//	return decodeJsonBody(resp, r)
	//}

	switch r.result.(type) {
	case string, *string:
		return decodeRawBody(resp, r)
	default:
		return decodeJsonBody(resp, r)
	}

}

// will support more types
func decodeJsonBody(resp *http.Response, r *Req) error {
	err := json.NewDecoder(resp.Body).Decode(r.result)
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if err != nil {
		return err
	}

	return nil
}

func decodeRawBody(resp *http.Response, r *Req) error {
	b, err := io.ReadAll(resp.Body)
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	if err != nil {
		return err
	}

	*(r.result.(*string)) = string(b)

	return nil
}
