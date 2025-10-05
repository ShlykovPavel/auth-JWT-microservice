package outbox_worker

import (
	"time"

	"github.com/go-co-op/gocron"
)

func SetupScheduler(worker *OutboxWorker) {
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(time.Second * 5).Do(func() {
		if err := worker.SendUsersToKafka(); err != nil {
			worker.logger.Error("error sending users to Kafka in scheduler:", err)
		}
	})
	scheduler.StartAsync()
	worker.logger.Info("Outbox worker scheduler started")
}
