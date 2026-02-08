package email

import (
	"context"

	"github.com/rs/zerolog"
)

type Sender interface {
	Send(ctx context.Context, to, subject, html string) error
}

type LogSender struct {
	Logger zerolog.Logger
}

func NewLogSender(logger zerolog.Logger) *LogSender {
	return &LogSender{Logger: logger}
}

func (s *LogSender) Send(_ context.Context, to, subject, html string) error {
	s.Logger.Info().
		Str("to", to).
		Str("subject", subject).
		Int("html_len", len(html)).
		Msg("email sent (log stub)")
	return nil
}
