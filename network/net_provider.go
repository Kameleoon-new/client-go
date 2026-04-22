package network

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpproxy"
)

// request

type HttpMethod string

const (
	HttpGet  HttpMethod = "GET"
	HttpPost HttpMethod = "POST"
)

type ContentType string

const (
	NoneContentType     ContentType = ""
	WildcardContentType ContentType = "*/*"
	TextContentType     ContentType = "text/plain"
	JsonContentType     ContentType = "application/json"
	FormContentType     ContentType = "application/x-www-form-urlencoded"
)

type Request struct {
	Method         HttpMethod        // mandatory
	Url            string            // mandatory
	ContentType    ContentType       // optional ("")
	Timeout        time.Duration     // mandatory
	Data           string            // optional ("")
	Headers        map[string]string // optional
	Authorization  string            // optional ("")
	IsAuthRequired bool              // optional (false)
}

func (r Request) String() string {
	body := "nil"
	if r.Data != "" {
		body = r.Data
	}

	var headersBuilder strings.Builder
	first := true
	for name, value := range r.Headers {
		if name == "Authorization" {
			value = "***"
		}
		if !first {
			headersBuilder.WriteString(",")
		}
		first = false

		headersBuilder.WriteString(name)
		headersBuilder.WriteString(":")
		headersBuilder.WriteString(value)
	}

	return fmt.Sprintf("HttpRequest{Method:'%s',Url:'%s',Headers:%v,Body:'%s'}", r.Method, r.Url, headersBuilder.String(), body)
}

// response

type Response struct {
	Err         error             // optional (nil)
	Code        int               // optional (0)
	Body        []byte            // optional ([])
	HeadersRead map[string]string // optional ({})
	Request     *Request          // mandatory
}

func (r Response) IsExpectedStatusCode() bool {
	return (r.Code/100 == 2) || (r.Code == 403) || (r.Code == 304)
}

func (r Response) String() string {
	return fmt.Sprintf("HttpResponse{Code:'%d',Reason:'%v',Body:'%s'}", r.Code, r.Err, string(r.Body))
}

// declaration

type NetProvider interface {
	Call(request *Request, headersToRead []string) Response
}

// implementation

const (
	AuthorizationHeader = "Authorization"
	MaxRedirectsCount   = 8
)

type NetProviderImpl struct {
	client fasthttp.Client
}

func NewNetProviderImpl(readTimeout time.Duration, writeTimeout time.Duration,
	maxConnsPerHost int, proxyUrl string) *NetProviderImpl {
	np := &NetProviderImpl{
		client: fasthttp.Client{
			ReadTimeout:     readTimeout,
			WriteTimeout:    writeTimeout,
			MaxConnsPerHost: maxConnsPerHost,
			TLSConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	if len(proxyUrl) > 0 {
		np.client.Dial = fasthttpproxy.FasthttpHTTPDialer(proxyUrl)
	}
	return np
}

func (np *NetProviderImpl) Call(request *Request, headersToRead []string) Response {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.Header.SetMethod(string(request.Method))
	req.Header.SetRequestURI(request.Url)
	np.setHeaders(req, request)
	if len(request.Data) > 0 {
		req.SetBodyString(request.Data)
	}
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	err := np.doTimeoutRedirects(req, resp, MaxRedirectsCount, request.Timeout)
	// err := np.client.DoTimeout(req, resp, request.Timeout)
	var response Response
	if err == nil {
		headersRead := np.readHeaders(resp, headersToRead)
		response = Response{Code: resp.StatusCode(), Body: resp.Body(), HeadersRead: headersRead, Request: request}
	} else {
		response = Response{Err: err, Request: request}
	}
	return response
}

func (*NetProviderImpl) setHeaders(req *fasthttp.Request, request *Request) {
	if len(request.Headers) > 0 {
		for key, value := range request.Headers {
			req.Header.Set(key, value)
		}
	}
	if len(request.Authorization) > 0 {
		switch {
		case strings.HasPrefix(request.Authorization, "Basic "):
			req.Header.Set(AuthorizationHeader, request.Authorization)
		default:
			req.Header.Set(AuthorizationHeader, "Bearer "+request.Authorization)
		}
	}
	if len(request.ContentType) > 0 {
		req.Header.SetContentType(string(request.ContentType))
	}
}

func (*NetProviderImpl) readHeaders(resp *fasthttp.Response, headersToRead []string) map[string]string {
	headers := make(map[string]string)
	for _, h := range headersToRead {
		headers[h] = string(resp.Header.Peek(h))
	}
	return headers
}

func (np *NetProviderImpl) doTimeoutRedirects(
	req *fasthttp.Request,
	resp *fasthttp.Response,
	redirectsLeft int,
	timeout time.Duration,
) error {
	deadline := time.Now().Add(timeout)
	for {
		if err := np.client.DoDeadline(req, resp, deadline); err != nil {
			return err
		}
		code := resp.StatusCode()
		if (code != 307) && (code != 308) {
			return nil
		}
		// redirection
		if redirectsLeft == 0 {
			return fasthttp.ErrTooManyRedirects
		}
		// supports only absolute urls
		loc := resp.Header.Peek("Location")
		if len(loc) == 0 {
			// redirection to nowhere
			return nil
		}
		req.URI().UpdateBytes(loc)
		redirectsLeft--
	}
}
