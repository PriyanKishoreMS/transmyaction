package jobs

import (
	"github.com/go-co-op/gocron/v2"
	"github.com/labstack/gommon/log"
	"github.com/priyankishorems/transmyaction/api/handlers"
)

func UpdateTxnsJob(h handlers.Handlers, scheduler gocron.Scheduler, atTimes gocron.AtTimes) (gocron.Job, error) {
	job, err := scheduler.NewJob(gocron.DailyJob(1, atTimes), gocron.NewTask(func() error {
		log.Info("Running updateTxnsJob")

		if err := h.UpdateTransactionsJob(); err != nil {
			return err
		}

		log.Info("updateTxnsJob completed")
		return nil
	}))

	return job, err
}
