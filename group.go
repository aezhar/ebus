// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
// Copyright(c) 2023 individual contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License
// that can be found in the LICENSE file.
// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *

package ebus

import (
	"slices"
	"sync"
)

// group represents a collection of subscriber instances.
type group[T Event] struct {
	m    sync.RWMutex
	subs []*subscriber[T]
}

// send sends an event to all consumers of the given event identified by its type.
func (g *group[T]) send(ev T) {
	g.m.RLock()
	defer g.m.RUnlock()

	for _, sub := range g.subs {
		sub.send(ev)
	}
}

// add registers the given handler as a subscriber to the group.
func (g *group[T]) add(qsize int, handler func(T)) *subscriber[T] {
	c := newConsumer[T](qsize)

	// Add the new subscriber to the list of active consumers.
	g.m.Lock()
	defer g.m.Unlock()
	g.subs = append(g.subs, c)

	// Start listening for events.
	go c.listen(handler)
	return c
}

// pop removes a subscriber from the group.
func (g *group[T]) pop(sub *subscriber[T]) {
	g.m.Lock()
	defer g.m.Unlock()

	// Search and remove the subscriber.
	if i := slices.Index(g.subs, sub); i >= 0 {
		sub.stop()

		g.subs[i] = g.subs[len(g.subs)-1]
		g.subs[len(g.subs)-1] = nil
		g.subs = g.subs[:len(g.subs)-1]
	}
}

// newGroup returns a new group instance.
func newGroup[T Event]() *group[T] {
	return &group[T]{}
}
