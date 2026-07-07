// Package correlate groups anomaly events into fault hypotheses using
// a rolling time window plus topology-distance grouping.
package correlate

import "time"

type Correlator struct {
	window time.Duration
}

func New(window time.Duration) *Correlator {
	return &Correlator{window: window}
}

func (c *Correlator) Window() time.Duration { return c.window }

// TODO: interval-tree keyed by topology distance; group anomalies
// that are graph-adjacent and time-adjacent into a single Fault.
