package job

import (
	"context"
	"encoding/json"
	"time"

	"github.com/hibiken/asynq"
)

const (
	QueueTypeCheckStatusTransaction = "transaction:check-status"
)

type CheckStatusTransactionPayload struct {
	TransactionID string    `json:"transaction_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type CheckStatusTransactionJob struct {
	Client *asynq.Client
}

func NewCheckTransactionJob(client *asynq.Client) CheckStatusTransactionJob {
	return CheckStatusTransactionJob{
		Client: client,
	}
}

func (j *CheckStatusTransactionJob) EnqueueCheckTransaction(ctx context.Context, transactionId string, transactionDuration time.Duration, timeout time.Duration) error {
	payload, err := json.Marshal(CheckStatusTransactionPayload{
		TransactionID: transactionId,
		CreatedAt:     time.Now(),
	})
	if err != nil {
		return err
	}

	task := asynq.NewTask(
		QueueTypeCheckStatusTransaction,
		payload,
		asynq.Timeout(timeout),
		asynq.ProcessIn(transactionDuration),
		asynq.MaxRetry(3),
	)

	// Enqueue ke queue default
	_, err = j.Client.EnqueueContext(ctx, task)
	return err
}
