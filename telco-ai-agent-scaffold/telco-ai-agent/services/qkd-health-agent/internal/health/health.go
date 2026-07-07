// Package health tracks per-link QKD channel health metrics and
// applies degradation thresholds (QBER, drift) to raise anomalies.
package health

type LinkHealth struct {
	LinkID           string
	QBER             float64
	SiftedKeyRateBps float64
	FinalKeyRateBps  float64
	SiftingRatio     float64
	ChannelDrift     float64
	Degraded         bool
}

type Monitor struct {
	links map[string]LinkHealth
}

func NewMonitor() *Monitor {
	return &Monitor{links: map[string]LinkHealth{}}
}

// Evaluate applies degradation thresholds; QBER above ~11% is the
// conventional BB84 security-abort threshold and is treated as
// Degraded regardless of other metrics.
func (m *Monitor) Evaluate(h LinkHealth) LinkHealth {
	if h.QBER > 0.11 {
		h.Degraded = true
	}
	m.links[h.LinkID] = h
	return h
}
