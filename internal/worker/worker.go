package worker

import (
	"time"

	"github.com/hibiken/asynq"
)

const (
	TypeScoreSnapshot     = "score:snapshot"
	TypeStreakUpdate       = "streak:update"
	TypeCounterReconcile  = "counter:reconcile"
	TypeSendNotification  = "notification:send"
	TypeInsightGenerate   = "insight:generate"
)

func NewTask(typeName string, payload []byte) *asynq.Task {
	return asynq.NewTask(typeName, payload, asynq.MaxRetry(3), asynq.Timeout(5*time.Minute))
}
