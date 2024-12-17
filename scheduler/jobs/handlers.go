package jobs

import (
	"context"
	elector "github.com/go-co-op/gocron-etcd-elector"
	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"log"
	"time"
)

type JobsRepository struct {
	scheduler *gocron.Scheduler
	elector   *elector.Elector
}

// CreateCronJob creates a cron job with the given scheduler, cron definition and job
// The definition of the cron job is in the format of a cron expression, example one every 10 seconds:
// "*/10 * * * * *"
func (cr *JobsRepository) CreateCronJob(cronDefinition string, ctx context.Context, job func(), jobUUID string) {
	scheduler := *cr.scheduler
	identifier, parseErr := uuid.Parse(jobUUID)
	if parseErr != nil {
		log.Printf("Failed to parse UUID: %v\n", parseErr)
	}
	_, err := scheduler.NewJob(gocron.CronJob(cronDefinition, true), gocron.NewTask(func() {
		if cr.elector.IsLeader(ctx) == nil {
			log.Printf("Running job %s\n", identifier)
			job()
		} else {
			log.Printf("Not leader, skipping job\n")
		}
	}), gocron.WithIdentifier(identifier))
	if err != nil {
		log.Printf("Failed to create cron job: %v\n", err)
	}
}

// CreateDelayedJob creates a delayed job with the given scheduler, delay in seconds and job
func (cr *JobsRepository) CreateDelayedJob(delay uint, ctx context.Context, job func(), jobUUID string) {
	scheduler := *cr.scheduler
	identifier, parseErr := uuid.Parse(jobUUID)
	if parseErr != nil {
		log.Printf("Failed to parse UUID: %v\n", parseErr)
	}
	_, err := scheduler.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(time.Now().Add(time.Duration(delay)*time.Second))),
		gocron.NewTask(func() {
			if cr.elector.IsLeader(ctx) == nil {
				job()
			} else {
				log.Printf("Not leader, skipping job\n")
			}
		}), gocron.WithIdentifier(identifier))
	if err != nil {
		log.Printf("Failed to create delayed job: %v\n", err)
	}
}

func (cr *JobsRepository) CancelJob(jobUUID string) {
	scheduler := *cr.scheduler
	identifier, parseErr := uuid.Parse(jobUUID)
	if parseErr != nil {
		log.Printf("Failed to parse UUID: %v\n", parseErr)
	}
	_ = scheduler.RemoveJob(identifier)
}
