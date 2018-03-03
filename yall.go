package yall

import (
	"context"
	"net/http"
	"time"
)

const (
	// Default is the level entries that don't specify
	// a severity get logged at.
	Default = Severity("DEFAULT")
	// Debug is useful for debugging, but is probably too noisy for standard
	// operation.
	Debug = Severity("DEBUG")
	// Info is routine information that does not indicate an error, but will
	// be helpful to non-debugging operators.
	Info = Severity("INFO")
	// Warning indicates a problem may have occurred, it is unclear. No
	// user-facing errors should occur from events that merit Warning, but
	// those events often predict a future error.
	Warning = Severity("WARN")
	// Error indicates a user-facing error occurred. Error logs that show up
	// repeatedly should lead to issue reports, which should lead to fixes.
	Error = Severity("ERROR")
)

type ctxKey struct{ string }

var logKey = struct{ string }{"log"}

// Severity represents a log severity. It's best to use one of the
// defined constants.
type Severity string

// Sink is something that can take log entries and persist them,
// either to a datastore or to the user's screen, or discard them.
type Sink interface {
	AddEntry(Entry)
	Flush() error
}

// Entry is a single log event, consisting of structured data and
// metadata.
type Entry struct {
	// LoggedAt is the time the entry was created. If zero, the current time is used.
	LoggedAt time.Time

	// Severity is the entry's severity level.
	// The zero value is Default.
	Severity Severity

	// Payload is a structured set of data to send as the log entry.
	Payload map[string]interface{}

	// Labels optionally specifies key/value metadata for the log entry.
	Labels map[string]string

	// HTTPRequest optionally specifies metadata about the HTTP request
	// associated with this log entry, if applicable. It is optional.
	HTTPRequest *HTTPRequest

	// Error specifies information about an error
	Error struct {
		// Error is the error being logged.
		Err error
		// Stacktrace is the stacktrace--including the error.
		Stacktrace string
	}

	// User specifies the user the request was made on behalf of.
	User string
}

// HTTPRequest defines metadata for an HTTP request associated with an
// entry.
type HTTPRequest struct {
	// Request is the http.Request passed to the handler.
	Request *http.Request

	// RequestSize is the size of the HTTP request message in bytes, including
	// the request headers and the request body.
	RequestSize int64

	// Status is the response code indicating the status of the response.
	// Examples: 200, 404.
	Status int

	// ResponseSize is the size of the HTTP response message sent back to the client, in bytes,
	// including the response headers and the response body.
	ResponseSize int64

	// Latency is the request processing latency on the server, from the time the request was
	// received until the response was sent.
	Latency time.Duration

	// LocalIP is the IP address (IPv4 or IPv6) of the origin server that the request
	// was sent to.
	LocalIP string

	// RemoteIP is the IP address (IPv4 or IPv6) of the client that issued the
	// HTTP request. Examples: "192.168.1.1", "FE80::0202:B3FF:FE1E:8329".
	RemoteIP string

	// CacheHit reports whether an entity was served from cache (with or without
	// validation).
	CacheHit bool

	// CacheValidatedWithOriginServer reports whether the response was
	// validated with the origin server before being served from cache. This
	// field is only meaningful if CacheHit is true.
	CacheValidatedWithOriginServer bool
}

func InContext(ctx context.Context, l *Logger) context.Context {
	return context.WithValue(ctx, logKey, l)
}

func FromContext(ctx context.Context) *Logger {
	l := ctx.Value(logKey)
	if l == nil {
		return &Logger{}
	}
	logger, ok := l.(*Logger)
	if !ok {
		return &Logger{}
	}
	return logger
}
