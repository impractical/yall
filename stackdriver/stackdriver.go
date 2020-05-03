package stackdriver

import (
	"cloud.google.com/go/logging"
	uuid "github.com/hashicorp/go-uuid"
	yall "yall.in"
)

func New(log *logging.Logger) Logger {
	return Logger{
		log: log,
	}
}

type Logger struct {
	log *logging.Logger
}

func (l Logger) AddEntry(e yall.Entry) {
	if l.log == nil {
		return
	}
	id, err := uuid.GenerateUUID()
	if err != nil {
		// we're explicitly ignoring an error
		// here because it can only occur
		// when the source of cryptographic
		// randomness on the machine can't be
		// read. At which point, what are we
		// going to do, anyways?
		return
	}
	l.log.Log(logging.Entry{
		Timestamp:   e.LoggedAt,
		Severity:    logging.ParseSeverity(string(e.Severity)),
		Payload:     e.Payload,
		Labels:      e.Labels,
		InsertID:    id,
		HTTPRequest: (*logging.HTTPRequest)(e.HTTPRequest),
		// TODO(paddy): maybe set Operation?
		//TODO(paddy): may be set Resource?
	})
}

func (l Logger) Flush() error {
	if l.log == nil {
		return nil
	}
	return l.log.Flush()
}
