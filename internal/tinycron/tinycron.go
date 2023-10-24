package tinycron

import (
	"context"
	"time"
)

// EveryDay is a non blocking function that runs f once a day at the top of the given hour [0-23], and also when first called
func EveryDay(ctx context.Context, hour int, f func()) {
	go func() {
		timer := time.NewTimer(time.Hour)
		defer timer.Stop()
		for {
			f()

			now := time.Now()

			// overwrite now with the given hour
			t := time.Date(now.Year(), now.Month(), now.Day(), hour, now.Minute(), now.Second(), now.Nanosecond(), now.Location())

			// if it was in the past, add a day
			if !t.After(now) {
				t = t.AddDate(0, 0, 1)
			}

			timer.Reset(time.Until(t))

			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				continue
			}
		}
	}()
}
