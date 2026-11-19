package streaming

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Kameleoon/client-go/v3/logging"
)

const (
	sseContentType = "text/event-stream"
	// sseConnectTimeout bounds each step of establishing a connection: the dial, the TLS
	// handshake and the wait for the response headers.
	sseConnectTimeout = 30 * time.Second
	// sseIdleTimeout detects stalled and half-dead connections (e.g. silently dropped by a NAT
	// or a load balancer), which would otherwise hang forever. Must stay above the server's
	// keep-alive interval (180 s, `server.sse.keepalive.writer-idle` in ms-client-configs), so
	// don't change it without the agreement of the ms-client-configs team.
	sseIdleTimeout = 6 * time.Minute
)

// errSseStreamStalled reports a stream which delivered nothing for `sseIdleTimeout`.
var errSseStreamStalled = errors.New("no data received within " + sseIdleTimeout.String())

// SseEvent is a single server-sent event.
type SseEvent struct {
	Name string // the `event` field, empty for an unnamed event
	Data string
}

// SseClient opens server-sent event streams.
type SseClient interface {
	// Connect returns an open stream, or an error when the connection could not be
	// established (network error, unexpected response). The stream ends once `ctx` is done.
	Connect(ctx context.Context, eventUrl string) (SseStream, error)
}

// SseStream is an open server-sent event stream.
type SseStream interface {
	// Next blocks until the next event. It returns `io.EOF` once the server has closed the
	// stream, and any other error when the stream has failed.
	Next() (SseEvent, error)
	Close()
}

// HttpSseClient is the `SseClient` over `net/http`.
type HttpSseClient struct {
	client      *http.Client
	idleTimeout time.Duration
}

// NewHttpSseClient creates a client which connects through `proxyUrl` (`host:port`, with an
// optional scheme and credentials) if it is set, and through the proxy configured by the
// environment (`HTTPS_PROXY`, ...) otherwise.
func NewHttpSseClient(proxyUrl string) *HttpSseClient {
	proxy := http.ProxyFromEnvironment
	if proxyUrl != "" {
		if !strings.Contains(proxyUrl, "://") {
			proxyUrl = "http://" + proxyUrl
		}
		if u, err := url.Parse(proxyUrl); err != nil {
			// `url.Error` repeats the whole URL, credentials included
			var urlErr *url.Error
			if errors.As(err, &urlErr) {
				err = urlErr.Err
			}
			logging.Error("Invalid proxy URL for the SSE connection: %s; using the environment settings", err)
		} else {
			proxy = http.ProxyURL(u)
		}
	}
	transport := &http.Transport{
		Proxy:                 proxy,
		DialContext:           (&net.Dialer{Timeout: sseConnectTimeout}).DialContext,
		TLSHandshakeTimeout:   sseConnectTimeout,
		ResponseHeaderTimeout: sseConnectTimeout,
	}
	return &HttpSseClient{client: &http.Client{Transport: transport}, idleTimeout: sseIdleTimeout}
}

func (c *HttpSseClient) Connect(ctx context.Context, eventUrl string) (SseStream, error) {
	// Each connection gets its own context, so the idle timer can abort the connection
	// without closing the whole service.
	ctx, cancel := context.WithCancel(ctx)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, eventUrl, nil)
	if err != nil {
		cancel()
		return nil, err
	}
	req.Header.Set("Accept", sseContentType)
	req.Header.Set("Cache-Control", "no-cache")
	resp, err := c.client.Do(req)
	if err != nil {
		cancel()
		return nil, err
	}
	if err := validateSseResponse(resp); err != nil {
		resp.Body.Close()
		cancel()
		return nil, err
	}
	return newHttpSseStream(cancel, resp.Body, c.idleTimeout), nil
}

func validateSseResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected response status: %s", resp.Status)
	}
	contentType := resp.Header.Get("Content-Type")
	if mediaType, _, _ := mime.ParseMediaType(contentType); mediaType != sseContentType {
		return fmt.Errorf("unexpected response content type: '%s'", contentType)
	}
	return nil
}

type httpSseStream struct {
	cancel      context.CancelFunc
	body        io.ReadCloser
	scanner     *bufio.Scanner
	firstLine   bool
	idleTimeout time.Duration
	idleTimer   *time.Timer
	stalled     int32 // set atomically by the idle timer
}

func newHttpSseStream(cancel context.CancelFunc, body io.ReadCloser, idleTimeout time.Duration) *httpSseStream {
	// The events are small (a few hundred bytes), so the scanner's default 64 KB line limit is
	// enough; a longer line fails the stream with `bufio.ErrTooLong`.
	scanner := bufio.NewScanner(body)
	scanner.Split(newSseLineSplitter())
	s := &httpSseStream{
		cancel:      cancel,
		body:        body,
		scanner:     scanner,
		firstLine:   true,
		idleTimeout: idleTimeout,
	}
	s.idleTimer = time.AfterFunc(idleTimeout, func() {
		atomic.StoreInt32(&s.stalled, 1)
		cancel() // fails the blocked read
	})
	return s
}

// newSseLineSplitter recognizes CRLF, lone CR and lone LF line endings. A CR ends the
// line immediately; an optional following LF is skipped even if it arrives in a later
// read. Each scanner needs its own splitter to retain that state. Partial lines at EOF
// are discarded.
func newSseLineSplitter() bufio.SplitFunc {
	skipLF := false
	return func(data []byte, _ bool) (advance int, token []byte, err error) {
		start := 0
		if skipLF && len(data) > 0 {
			skipLF = false
			if data[0] == '\n' {
				start = 1
			}
		}
		for i := start; i < len(data); i++ {
			if data[i] == '\r' || data[i] == '\n' {
				skipLF = data[i] == '\r'
				return i + 1, data[start:i], nil
			}
		}
		return start, nil, nil
	}
}

// Next parses the stream as per the SSE specification: an event is dispatched by an empty
// line, `data` lines are joined with newlines, comment lines (keep-alives) are skipped, and
// an empty line without data discards the pending event name.
func (s *httpSseStream) Next() (SseEvent, error) {
	var evt SseEvent
	var data strings.Builder
	hasData := false
	for {
		if !s.scanner.Scan() {
			if atomic.LoadInt32(&s.stalled) != 0 {
				return SseEvent{}, errSseStreamStalled
			}
			if err := s.scanner.Err(); err != nil {
				return SseEvent{}, err
			}
			return SseEvent{}, io.EOF
		}
		s.idleTimer.Reset(s.idleTimeout)
		line := s.scanner.Text()
		if s.firstLine {
			s.firstLine = false
			line = strings.TrimPrefix(line, "\uFEFF") // a leading byte order mark is ignored
		}
		if line == "" {
			if hasData {
				evt.Data = data.String()
				return evt, nil
			}
			evt = SseEvent{}
			continue
		}
		if line[0] == ':' {
			continue
		}
		field, value := line, ""
		if i := strings.IndexByte(line, ':'); i >= 0 {
			field, value = line[:i], strings.TrimPrefix(line[i+1:], " ")
		}
		switch field {
		case "event":
			evt.Name = value
		case "data":
			if hasData {
				data.WriteByte('\n')
			}
			data.WriteString(value)
			hasData = true
		}
		// `id` and `retry` are not used by the SDK
	}
}

func (s *httpSseStream) Close() {
	s.idleTimer.Stop()
	s.cancel()
	s.body.Close()
}
