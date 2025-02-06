package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const (
	HTTPS = "https://"
	HTTP  = "http://"
)

type option func(request *http.Request)

type post func(response *http.Response)

type Req struct {
	req    *http.Request
	method string
	rawURL string
	body   interface{}
	header interface{}
	query  interface{}

	retryStrategy func(*http.Response, error) bool

	ctx    context.Context
	cancel context.CancelFunc

	process []option
	posts   []post

	result interface{}
}

func (r *Req) String() string {
	trackingId, _ := r.ctx.Value("rid").(string)

	return fmt.Sprintf(
		"trackingId: %s, method: %s, url: %s",
		trackingId,
		r.method,
		r.rawURL,
	)
}

func NewReq(method string, rawUrl string) *Req {
	if !strings.HasPrefix(rawUrl, HTTP) {
		rawUrl = HTTP + rawUrl
	}

	return &Req{method: method, rawURL: rawUrl, ctx: context.Background()}
}

func NewReqTLS(method string, rawUrl string) *Req {
	if !strings.HasPrefix(rawUrl, HTTPS) {
		rawUrl = HTTPS + rawUrl
	}

	return &Req{method: method, rawURL: rawUrl, ctx: context.Background()}
}

func (r *Req) SetContext(ctx context.Context) *Req {
	r.ctx = ctx
	return r
}

func (r *Req) SetHeader(h interface{}) *Req {
	r.header = h
	return r
}

func (r *Req) SetQuery(query interface{}) *Req {
	r.query = query
	return r
}

func (r *Req) SetBody(encoder interface{}) *Req {
	r.body = encoder
	return r
}

func (r *Req) BindBody(body interface{}) *Req {
	r.result = body
	return r
}

func (r *Req) SetRetryStrategy(retryStrategy func(response *http.Response, err error) bool) *Req {
	r.retryStrategy = retryStrategy
	return r
}

func DefaultRetryStrategy(response *http.Response, err error) bool {
	if err != nil {
		return true
	}

	if response == nil {
		return true
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return true
	}

	return false
}

func (r *Req) New() (*http.Request, error) {

	var (
		body []byte
		req  *http.Request
		err  error
	)
	if r.body != nil {
		body, err = encodeBody(r.body)
		if err != nil {
			return nil, err
		}
	}

	var v = url.Values{}
	if r.query != nil {
		encode(r.query, encoder{adder: v, tag: "query"})
	}

	req, err = http.NewRequestWithContext(r.ctx, r.method, r.rawURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = v.Encode()

	if r.header != nil {
		encode(r.header, encoder{adder: req.Header, tag: "header"})
	}

	for _, process := range r.process {
		process(req)
	}

	return req, nil
}
