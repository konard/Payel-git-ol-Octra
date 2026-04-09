package ratelimit

func (rl *RateLimiter) addBucket(name string, l limit) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.buckets[name] = newTokenBucket(l.max, l.window)
}
