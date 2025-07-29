package api

import (
	"assist-tix/config"
	"assist-tix/internal/job"

	"github.com/hibiken/asynq"
)

type Job struct {
	CheckStatusTransactionJob job.CheckStatusTransactionJob
}

func NewJob(
	env *config.EnvironmentVariable,
	asynqClient *asynq.Client,
) Job {
	checkStatusTransactionJob := job.NewCheckTransactionJob(asynqClient)
	return Job{
		CheckStatusTransactionJob: checkStatusTransactionJob,
	}
}
