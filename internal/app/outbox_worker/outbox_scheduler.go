package outbox_worker

import (
	"time"

	"github.com/go-co-op/gocron"
)

func SetupScheduler(worker *OutboxWorker, interval time.Duration) *gocron.Scheduler {
	scheduler := gocron.NewScheduler(time.UTC)
	// Джоба отправки пользователей в кафку
	scheduler.Every(interval).Do(func() {
		if err := worker.SendUsersToKafka(); err != nil {
			worker.logger.Error("error sending users to Kafka in scheduler", "error", err.Error())
		}
	})
	//джоба отметки записей как failed если превышен лимит попыток
	scheduler.Every(interval).Do(func() {
		if err := worker.MarkFailedUsers(); err != nil {
			worker.logger.Error("error marking records as failed in scheduler", "error", err.Error())
		}
	})
	scheduler.StartAsync()
	worker.logger.Info("Outbox worker scheduler started")
	return scheduler
}
