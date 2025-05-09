package roundrobin

import (
	"sync/atomic"

	"github.com/kurochkinivan/load_balancer/internal/entity"
)

type RobinRound struct {
	n        int32
	backends []*entity.Backend
	current  atomic.Int32
}

// New creates a new RobinRound balancer.
func New(backends []*entity.Backend) *RobinRound {
	n := len(backends)
	return &RobinRound{
		n:        int32(n),
		backends: backends,
	}
}

// Next returns the index of the next available backend.
//
// It is concurrently safe.
func (r *RobinRound) Next() (int32, bool) {
	var count int32
	for ; count != r.n; count++ {
		current := r.current.Add(1)
		idx := (current - 1) % r.n

		if r.backends[idx].IsAvailable() {
			return idx, true
		}
	}
	return -1, false
}

// Reset resets the balancer to its initial state.
//
// It is concurrently safe.
func (r *RobinRound) Reset() {
	r.current.Swap(0)
}
