package testing

import (
	"fmt"
	"strings"
	"testing"

	testInterface "github.com/mitchellh/go-testing-interface"
	yall "yall.in"
)

type Logger struct {
	t   *testing.T
	sev yall.Severity
}

func (l Logger) shouldLog(sev yall.Severity) bool {
	switch {
	case l.sev == yall.Error && sev != yall.Error:
		return false
	case l.sev == yall.Warning && sev != yall.Error && sev != yall.Warning:
		return false
	case l.sev == yall.Info && sev != yall.Error && sev != yall.Warning && sev != yall.Info:
		return false
	default:
		// when logging at debug level log default & unknown sevs, too
		return true
	}
}

func New(t *testing.T, sev yall.Severity) Logger {
	return Logger{
		t:   t,
		sev: sev,
	}
}

type sprintf func(string, ...interface{}) string

func (l Logger) AddEntry(e yall.Entry) {
	if l.t == nil {
		return
	}
	l.t.Helper()
	if !l.shouldLog(e.Severity) {
		return
	}
	var msg string
	if v, ok := e.Payload["msg"]; ok {
		msg = fmt.Sprintf(" %v", v)
	}

	var fields []string
	for k, v := range e.Payload {
		if k == "msg" {
			continue
		}
		fields = append(fields, fmt.Sprintf("%s=%v", k, v))
	}

	l.t.Logf("%s [%s]%s\n", e.LoggedAt.Local().Format("2006/01/02 15:04:05.0000"), e.Severity, msg)
	if e.Error.Err != nil {
		l.t.Logf("\t\terr=%s\n", e.Error.Err.Error())
	}
	if len(fields) > 0 {
		l.t.Logf("\t\t%s\n", strings.Join(fields, "\t"))
	}
	if e.Error.Stacktrace != "" {
		l.t.Logf("\nStacktrace:\n%s\n", e.Error.Stacktrace)
	}
}

func (l Logger) Flush() error {
	return nil
}

func (l Logger) TestHelper() testInterface.T {
	return l.t
}

var _ yall.SinkWithTestHelper = Logger{}
