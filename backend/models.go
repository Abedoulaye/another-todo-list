package main

import (
	"sync"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type Note struct {
	ID        int64     `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type rateLimiter struct {
	mu       sync.Mutex          // used to protect shared data from being accessed by multiple goroutines at the same time. if you  try to read/write the visitors map at the same time, you get a data race — which can crash your program or corrupt data. A mutex lets you say "only one goroutine at a time can touch this."
	visitors map[string]*visitor // a map of ips and their count, the string key is the ip part
}

type visitor struct {
	count   int
	resetAt time.Time
}
