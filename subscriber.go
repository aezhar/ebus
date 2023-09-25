// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
// Copyright(c) 2023 individual contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License
// that can be found in the LICENSE file.
// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *

package ebus

// subscriber represents a single subscriber with a message queue.
type subscriber[T Event] struct {
	queue chan T
	done  chan struct{}
}

// listen listens to the event queue and processes events.
func (s *subscriber[T]) listen(cb func(T)) {
	defer close(s.done)
	for ev := range s.queue {
		cb(ev)
	}
}

// send queues an event for the listener to process.
func (s *subscriber[T]) send(ev T) {
	s.queue <- ev
}

// stop closes the event queue and waits for the listener to complete.
func (s *subscriber[T]) stop() {
	close(s.queue)
	<-s.done
}

// newConsumer creates a new consumer.
func newConsumer[T Event](qsize int) *subscriber[T] {
	return &subscriber[T]{
		queue: make(chan T, qsize),
		done:  make(chan struct{}),
	}
}
