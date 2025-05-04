package entity

import "sync/atomic"

type Client struct {
	ID            int64  `json:"id"`
	IPAddress     string `json:"ip_address"`
	Name          string `json:"name"`
	Capacity      int32  `json:"capacity"`
	RatePerSecond int32  `json:"rate_per_second"`
	Tokens        atomic.Int32
}

// Allow checks if client has available tokens.
//
// This method is concurrently safe.
func (c *Client) Allow() bool {
	for {
		current := c.Tokens.Load()
		if current == 0 {
			return false
		}
		if c.Tokens.CompareAndSwap(current, current-1) {
			return true
		}
	}
}

// RefillTokensOncePerSecond refills client tokens by RatePerSecond.
// It should be called once per second
//
// This method is concurrently safe.
func (c *Client) RefillTokensOncePerSecond() {
	for {
		current := c.Tokens.Load()
		newTokens := min(current+int32(c.RatePerSecond), int32(c.Capacity))

		if c.Tokens.CompareAndSwap(current, newTokens) {
			break
		}
	}
}
