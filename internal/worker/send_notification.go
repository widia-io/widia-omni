package worker

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/widia-io/widia-omni/internal/email"
)

type SendNotificationPayload struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

type SendNotificationHandler struct {
	emailSender email.Sender
	logger      zerolog.Logger
}

func NewSendNotificationHandler(emailSender email.Sender, logger zerolog.Logger) *SendNotificationHandler {
	return &SendNotificationHandler{emailSender: emailSender, logger: logger}
}

func (h *SendNotificationHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p SendNotificationPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	if err := h.emailSender.Send(ctx, p.Email, p.Subject, p.HTML); err != nil {
		h.logger.Error().Err(err).Str("to", p.Email).Msg("failed to send email")
		return err
	}

	h.logger.Info().Str("to", p.Email).Str("subject", p.Subject).Msg("email sent")
	return nil
}
