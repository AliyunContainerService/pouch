package scheduler

import (
	"context"
	"fmt"
)

// Scheduler is an interface that implement a function
// choosing an item from a items pool.
type Scheduler interface {
	// Return a chosed client
	Schedule(ctx context.Context) (Factory, error)
}

// Factory is an interface that can produce and consume
// "goods"
type Factory interface {
	// Value will return the total numbers of goods
	Value() int

	// Produce will produce numbers of goods
	Produce(goods int)

	// Consume will consume numbers of goods
	Consume(goods int) error
}

// LRUScheduler is a Least Recently Used scheduler.
type LRUScheduler struct {
	pool []Factory
}

// NewLRUScheduler new a LRU scheduler.
func NewLRUScheduler(pool []Factory) (Scheduler, error) {
	return &LRUScheduler{
		pool: pool,
	}, nil
}

// Schedule is to choose the next candidate.
func (lru *LRUScheduler) Schedule(ctx context.Context) (Factory, error) {
	if len(lru.pool) == 0 {
		return nil, fmt.Errorf("empty candidate list")
	}

	var (
		index = 0
		least = lru.pool[0].Value()
	)

	for i := 1; i < len(lru.pool); i++ {
		v := lru.pool[i].Value()
		if v > least {
			index = i
			least = v
		}
	}

	// the max number of goods below 0, resources
	// have reached the limit.
	if least <= 0 {
		return nil, fmt.Errorf("resources exhausted")
	}

	return lru.pool[index], nil
}
