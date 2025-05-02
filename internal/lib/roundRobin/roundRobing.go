package roundrobin

import "sync/atomic"

type RobinRound struct {
	n       int32
	current atomic.Int32
}

func New(n int) *RobinRound {
	if n <= 0 {
		panic("n should be greater that 0")
	}
	return &RobinRound{n: int32(n)}
}

func (r *RobinRound) Next() int32 {
	current := r.current.Add(1)
	return (current - 1) % r.n
}

func (r *RobinRound) Reset() {
	r.current.Swap(0)
}
