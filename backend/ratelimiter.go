package main

import (
	"time"
)

func (rl *rateLimiter) allow(ip string, limit int, window time.Duration) bool {
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]

	if !exists || now.After(v.resetAt) {
		rl.visitors[ip] = &visitor{
			count:   1,
			resetAt: now.Add(window),
		}
		return true
	}

	if v.count == limit {
		return false
	}

	v.count++
	return true
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.After(v.resetAt) {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}
