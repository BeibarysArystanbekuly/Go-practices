package worker

import (
	"context"
	"log"
)

type VoteEvent struct {
	PollID   int64
	OptionID int64
}

type StatsWorker struct {
	Ch <-chan VoteEvent
}

func NewStatsWorker(ch <-chan VoteEvent) *StatsWorker {
	return &StatsWorker{Ch: ch}
}

func (w *StatsWorker) Run(ctx context.Context) {
	log.Println("stats worker started")
	for {
		select {
		case <-ctx.Done():
			log.Println("stats worker stopped")
			return
		case ev := <-w.Ch:
			// здесь просто логируем событие; при желании позже добавим обновление агрегатов в БД
			log.Printf("processing vote event: poll=%d option=%d\n", ev.PollID, ev.OptionID)
		}
	}
}
