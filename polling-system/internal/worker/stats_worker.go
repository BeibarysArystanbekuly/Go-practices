package worker

import (
	"context"
	"log"
	"sync"
	"time"
)

type VoteEvent struct {
	PollID   int64
	OptionID int64
}

type Aggregator interface {
	IncrementAggregated(ctx context.Context, pollID, optionID int64) error
}

type StatsWorker struct {
	Ch         <-chan VoteEvent
	agg        Aggregator
	workers    int
	retryDelay time.Duration
}

func NewStatsWorker(ch <-chan VoteEvent, agg Aggregator) *StatsWorker {
	return &StatsWorker{
		Ch:         ch,
		agg:        agg,
		workers:    2,
		retryDelay: 150 * time.Millisecond,
	}
}

func (w *StatsWorker) Run(ctx context.Context) {
	log.Println("stats worker pool started")
	var wg sync.WaitGroup
	for i := 0; i < w.workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			w.loop(ctx, id)
		}(i)
	}

	<-ctx.Done()
	wg.Wait()
	log.Println("stats worker pool stopped")
}

func (w *StatsWorker) loop(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-w.Ch:
			w.process(ctx, workerID, ev)
		}
	}
}

func (w *StatsWorker) process(ctx context.Context, workerID int, ev VoteEvent) {
	for attempt := 0; attempt < 3; attempt++ {
		if err := w.agg.IncrementAggregated(ctx, ev.PollID, ev.OptionID); err == nil {
			log.Printf("worker %d aggregated poll=%d option=%d", workerID, ev.PollID, ev.OptionID)
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(w.retryDelay * time.Duration(attempt+1)):
		}
	}
	log.Printf("worker %d failed to aggregate poll=%d option=%d after retries", workerID, ev.PollID, ev.OptionID)
}
