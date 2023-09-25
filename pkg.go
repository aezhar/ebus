// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
// Copyright(c) 2023 individual contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License
// that can be found in the LICENSE file.
// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *

package ebus

// StopFn is called to stop receiving any further events.
type StopFn func()

type EventType interface {
	Hash() uintptr
	Equals(other EventType) bool
}

// Event is a message on the Bus identified by an EventType value.
type Event interface {
	// Type always returns the same EventType for the given Event type.
	Type() EventType
}

// On subscribes a given handler to a given event type on the given bus.
func On[T Event](bus *Bus, cb func(T)) StopFn {
	var ev T
	return SubscribeTo(bus, 1, ev.Type(), cb)
}

// SubscribeTo subscribes to an event with the specified event type.
func SubscribeTo[T Event](bus *Bus, qsize int, et EventType, cb func(T)) StopFn {
	o, _ := bus.groups.GetOrCompute(et, func() any {
		return newGroup[T]()
	})

	g := o.(*group[T])

	s := g.add(qsize, cb)

	return func() { g.pop(s) }
}

// Emit sends the given event to all subscribers of the given event type.
func Emit[T Event](bus *Bus, ev T) {
	if o, ok := bus.groups.Get(ev.Type()); ok {
		o.(*group[T]).send(ev)
	}
}
