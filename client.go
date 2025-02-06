package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"time"
)

type Client struct {
	cli        *http.Client
	attempt    int
	curAttempt int

	wait    time.Duration
	maxWait time.Duration

	log log.Logger
}

// ClientOption is the option of the client
type ClientOption func(client *Client)

func WithMaxWait(maxWait time.Duration) ClientOption {
	return func(client *Client) {
		client.maxWait = maxWait
	}
}

func WithAttempt(attempt int) ClientOption {
	return func(client *Client) {
		client.attempt = attempt
	}
}

func WithWait(wait time.Duration) ClientOption {
	return func(client *Client) {
		client.wait = wait
	}
}

func WithTimeout(timeout time.Duration) ClientOption {
	return func(client *Client) {
		client.cli.Timeout = timeout
	}
}

func WithHttpClient(httpClient *http.Client) ClientOption {
	return func(client *Client) {
		client.cli = httpClient
	}
}

// WithTLSCert replace the ca cert of the http client.
func WithTLSCert(tlsConf *tls.Config) ClientOption {

	transport := &http.Transport{
		TLSClientConfig: tlsConf,
	}

	return func(client *Client) {

		if client.cli == nil {
			client.cli = &http.Client{
				Transport: transport,
			}
		}

		client.cli.Transport.(*http.Transport).TLSClientConfig = tlsConf
	}
}

func WithLogger(l log.Logger) ClientOption {
	return func(client *Client) {
		client.log = l
	}
}

func Retry(opts ...ClientOption) *Client {

	c := &Client{
		cli: &http.Client{
			Transport: http.DefaultTransport,
			Timeout:   time.Second,
		},

		attempt:    1,
		curAttempt: 0,
		wait:       time.Millisecond * 1000,
		maxWait:    time.Second * 30,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Do send the request
func (c *Client) Do(request *Req) (response *http.Response, err error) {

	var body []byte
	var r *http.Request
	for ; c.curAttempt < c.attempt; c.curAttempt++ {
		r, err = request.New()
		if err != nil {
			break
		}

		response, err = c.cli.Do(r)

		body = nil
		if response != nil {
			body, _ = io.ReadAll(response.Body)
			_ = response.Body.Close()
			response.Body = io.NopCloser(bytes.NewReader(body))
		}

		var retry bool
		if request.retryStrategy != nil {
			retry = request.retryStrategy(response, err)
		}

		if c.curAttempt++; !retry || c.curAttempt >= c.attempt {
			break
		}

		ctx := request.ctx
		if request.ctx == nil {
			ctx = context.Background()
		}

		// there's an interval time between each retry. The interval time is calculated by getWait()
		// and if the interval time is greater than maxWait, the interval time will be set to maxWait
		// and the interval time will be doubled each time
		// the interval time is random
		// If the set overall time-consuming limit is reached, no retry will be performed
		select {
		case <-time.After(c.getWait()):
		case <-ctx.Done():
			if request.cancel != nil {
				request.cancel()
			}
			break
		}

	}

	if response != nil {
		if len(body) > 0 {
			response.Body = io.NopCloser(bytes.NewReader(body))
		}

		err = decodeBody(response, request)
		if err != nil {
			return nil, err
		}

		for _, process := range request.posts {
			process(response)
		}
	}

	return
}

func (c *Client) getWait() time.Duration {
	temp := int64(c.wait * time.Duration(math.Exp2(float64(c.curAttempt-1))))
	if temp <= 0 {
		temp = int64(c.wait)
	}

	if temp > int64(c.maxWait) && c.maxWait != 0 {
		temp = int64(c.maxWait)
	}

	temp /= 2
	return time.Duration(temp) + time.Duration(rand.Int63n(temp))
}
