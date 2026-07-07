// Package registry tracks liveness and last-known state of every
// subagent (Topology, Anomaly, Fault Corr, QKD Health).
package registry

import "sync"

type AgentStatus struct {
	Name       string
	Alive      bool
	LastSeenMs int64
}

type Registry struct {
	mu     sync.RWMutex
	agents map[string]AgentStatus
}

func New() *Registry {
	return &Registry{agents: make(map[string]AgentStatus)}
}

func (r *Registry) Upsert(s AgentStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[s.Name] = s
}

func (r *Registry) Snapshot() []AgentStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]AgentStatus, 0, len(r.agents))
	for _, s := range r.agents {
		out = append(out, s)
	}
	return out
}
