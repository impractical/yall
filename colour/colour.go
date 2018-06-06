package colour

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"sync"

	"github.com/fatih/color"
	yall "yall.in"
)

var colors = map[string]*color.Color{
	"DEBUG": color.New(color.FgBlue),
	"INFO":  color.New(color.FgGreen),
	"WARN":  color.New(color.FgYellow),
	"ERROR": color.New(color.FgRed),
}

type Logger struct {
	mu  *sync.Mutex
	out io.Writer
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

func New(out io.Writer, sev yall.Severity) Logger {
	var m sync.Mutex
	return Logger{
		out: out,
		mu:  &m,
		sev: sev,
	}
}

type sprintf func(string, ...interface{}) string

func (l Logger) AddEntry(e yall.Entry) {
	if l.out == nil {
		return
	}
	if !l.shouldLog(e.Severity) {
		return
	}
	var sprinter sprintf
	sprinter = fmt.Sprintf
	if c, ok := colors[string(e.Severity)]; ok {
		sprinter = c.Sprintf
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
		fields = append(fields, sprinter("%s", k)+fmt.Sprintf("=%v", v))
	}
	sort.Strings(fields)

	l.mu.Lock()
	fmt.Fprintf(l.out, "%s %s%s\n", e.LoggedAt.Local().Format("2006/01/02 15:04:05.0000"), sprinter("[%s]", e.Severity), msg)
	if e.Error.Err != nil {
		fmt.Fprintf(l.out, "\t\t"+sprinter("err")+fmt.Sprintf("=%s\n", e.Error.Err.Error()))
	}
	if len(fields) > 0 {
		fmt.Fprintf(l.out, "\t\t%s\n", strings.Join(fields, "\t"))
	}
	if e.Error.Stacktrace != "" {
		fmt.Fprintf(l.out, "\nStacktrace:\n%s\n", e.Error.Stacktrace)
	}
	l.mu.Unlock()
}

func (l Logger) Flush() error {
	return nil
}
