package yall

import (
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

type Logger struct {
	sink      Sink
	payload   map[string]interface{}
	labels    map[string]string
	req       HTTPRequest
	reqStart  time.Time
	err       error
	stack     string
	user      string
	localIP   string
	proxyHops int
}

func New(s Sink) *Logger {
	return &Logger{
		sink: s,
	}
}

func (l *Logger) clone() *Logger {
	n := &Logger{
		sink:      l.sink,
		reqStart:  l.reqStart,
		err:       l.err,
		stack:     l.stack,
		user:      l.user,
		localIP:   l.localIP,
		proxyHops: l.proxyHops,
		req:       l.req,
		payload:   map[string]interface{}{},
		labels:    map[string]string{},
	}
	for k, v := range l.payload {
		n.payload[k] = v
	}
	for k, v := range l.labels {
		n.labels[k] = v
	}
	return n
}

func (l *Logger) WithProxyHops(n int) *Logger {
	log := l.clone()
	log.proxyHops = n
	return log
}

func (l *Logger) WithLocalIP(ip string) *Logger {
	log := l.clone()
	log.localIP = ip
	return log
}

func (l *Logger) WithField(k string, v interface{}) *Logger {
	log := l.clone()
	log.payload[k] = v
	return log
}

func (l *Logger) WithLabel(k, v string) *Logger {
	log := l.clone()
	log.labels[k] = v
	return log
}

func (l *Logger) WithRequest(r *http.Request) *Logger {
	log := l.clone()
	log.req.Request = r
	log.req.LocalIP = log.localIP
	log.req.RemoteIP = strings.Split(r.RemoteAddr, ":")[0]
	if r.Header.Get("X-Real-Ip") != "" {
		ips := strings.Split(r.Header.Get("X-Real-Ip"), ",")
		if len(ips) > log.proxyHops+1 {
			ips = ips[:len(ips)-log.proxyHops]
			log.req.RemoteIP = ips[len(ips)-1]
		}
	}
	if r.Header.Get("X-Forwarded-For") != "" {
		ips := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
		if len(ips) > log.proxyHops+1 {
			ips = ips[:len(ips)-log.proxyHops]
			log.req.RemoteIP = ips[len(ips)-1]
		}
	}
	strings.TrimSpace(log.req.RemoteIP)
	return log
}

func (l *Logger) StartRequest() *Logger {
	log := l.clone()
	log.reqStart = time.Now()
	return log
}

func (l *Logger) EndRequest(status int, responseSize int64) *Logger {
	log := l.clone()
	log.req.Latency = time.Now().Sub(l.reqStart)
	log.req.Status = status
	log.req.ResponseSize = responseSize
	return log
}

func (l *Logger) WithUser(id string) *Logger {
	log := l.clone()
	log.user = id
	return log
}

func (l *Logger) WithError(err error) *Logger {
	log := l.clone()
	log.err = err
	log.stack = string(debug.Stack())
	return log
}

func (l *Logger) Entry(sev Severity) Entry {
	e := Entry{
		LoggedAt: time.Now(),
		Severity: sev,
		Payload:  l.payload,
		Labels:   l.labels,
		User:     l.user,
	}
	if l.req.Request != nil || l.req.Latency != 0 {
		e.HTTPRequest = &l.req
	}
	if l.stack != "" || l.err != nil {
		e.Error.Err = l.err
		e.Error.Stacktrace = l.stack
	}
	return e
}

func (l *Logger) Error(msg string) {
	if l.sink == nil {
		return
	}
	if rh, ok := l.sink.(SinkWithTestHelper); ok {
		rh.TestHelper().Helper()
	}
	l.sink.AddEntry(l.WithField("msg", msg).Entry(Error))
}

func (l *Logger) Warn(msg string) {
	if l.sink == nil {
		return
	}
	if rh, ok := l.sink.(SinkWithTestHelper); ok {
		rh.TestHelper().Helper()
	}
	l.sink.AddEntry(l.WithField("msg", msg).Entry(Warning))
}

func (l *Logger) Debug(msg string) {
	if l.sink == nil {
		return
	}
	if rh, ok := l.sink.(SinkWithTestHelper); ok {
		rh.TestHelper().Helper()
	}
	l.sink.AddEntry(l.WithField("msg", msg).Entry(Debug))
}

func (l *Logger) Info(msg string) {
	if l.sink == nil {
		return
	}
	if rh, ok := l.sink.(SinkWithTestHelper); ok {
		rh.TestHelper().Helper()
	}
	l.sink.AddEntry(l.WithField("msg", msg).Entry(Info))
}
