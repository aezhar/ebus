// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
// Copyright(c) 2023 individual contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License
// that can be found in the LICENSE file.
// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *

package ebus_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/OneOfOne/xxhash"
	"github.com/aezhar/ebus"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

type EventTypeIntA int

func (e EventTypeIntA) Hash() uintptr                { return uintptr(e) }
func (e EventTypeIntA) Equals(r ebus.EventType) bool { return e == r }

const TypeEvent1 EventTypeIntA = 0

type EventTypeIntB int

func (e EventTypeIntB) Hash() uintptr                { return uintptr(e) }
func (e EventTypeIntB) Equals(r ebus.EventType) bool { return e == r }

const TypeEvent2 EventTypeIntB = 0

type MyIntEvent1 struct{ Number int }

func (e MyIntEvent1) Type() ebus.EventType { return TypeEvent1 }

type MyIntEvent2 struct{ Text string }

func (e MyIntEvent2) Type() ebus.EventType { return TypeEvent2 }

type EventTypeSlug struct{ slug string }

func (e EventTypeSlug) Hash() uintptr                { return uintptr(xxhash.ChecksumString64(e.slug)) }
func (e EventTypeSlug) Equals(r ebus.EventType) bool { return e == r }

var (
	TypeEventA = EventTypeSlug{"A"}
	TypeEventB = EventTypeSlug{"B"}
)

type MySlugEventA struct{ Number int }

func (e MySlugEventA) Type() ebus.EventType { return TypeEventA }

type MySlugEventB struct{ Text string }

func (e MySlugEventB) Type() ebus.EventType { return TypeEventB }

func TestPublish(t *testing.T) {
	b := ebus.New()
	var wg sync.WaitGroup

	t.Run("int", func(t *testing.T) {
		// Subscribe, must be received in order
		var count int64
		defer ebus.On(b, func(ev MyIntEvent1) {
			defer wg.Done()
			assert.Equal(t, int(atomic.AddInt64(&count, 1)), ev.Number)
		})()

		// Publish
		wg.Add(3)
		ebus.Emit(b, MyIntEvent1{Number: 1})
		ebus.Emit(b, MyIntEvent1{Number: 2})
		ebus.Emit(b, MyIntEvent1{Number: 3})

		ebus.Emit(b, MyIntEvent2{Text: "foobar"})

		// Wait and check
		wg.Wait()
		assert.Equal(t, int64(3), count)
	})

	t.Run("slug", func(t *testing.T) {
		// Subscribe, must be received in order
		var count int64
		defer ebus.On(b, func(ev MySlugEventA) {
			defer wg.Done()
			assert.Equal(t, int(atomic.AddInt64(&count, 1)), ev.Number)
		})()

		// Publish
		wg.Add(3)
		ebus.Emit(b, MySlugEventA{Number: 1})
		ebus.Emit(b, MySlugEventA{Number: 2})
		ebus.Emit(b, MySlugEventA{Number: 3})

		ebus.Emit(b, MySlugEventB{Text: "foobar"})

		// Wait and check
		wg.Wait()
		assert.Equal(t, int64(3), count)
	})
}

func TestUnsubscribe(t *testing.T) {
	b := ebus.New()

	var wg sync.WaitGroup
	var count int64

	stop := ebus.On(b, func(ev MyIntEvent1) {
		defer wg.Done()
		atomic.AddInt64(&count, 1)
	})

	wg.Add(1)
	ebus.Emit(b, MyIntEvent1{Number: 1})
	wg.Wait()

	assert.Equal(t, int64(1), count)

	stop()

	goleak.VerifyNone(t)

	ebus.Emit(b, MyIntEvent1{Number: 2})
	assert.Equal(t, int64(1), count)

	// Call stop twice.
	stop()
}

func TestSubscribeDifferentType(t *testing.T) {
	b := ebus.New()

	ebus.SubscribeTo(b, 1, TypeEvent1, func(ev MyIntEvent1) {})

	assert.Panics(t, func() {
		ebus.SubscribeTo(b, 1, TypeEvent1, func(ev MyIntEvent2) {})
	})
}

func TestConcurrent(t *testing.T) {
	const amount = 100000
	const workers = 10
	var count int64
	var wg1, wg2 sync.WaitGroup

	wg2.Add(1)

	b := ebus.New()
	defer ebus.On(b, func(ev MyIntEvent1) {
		if current := atomic.AddInt64(&count, 1); current == amount*workers {
			wg2.Done()
		}
	})()

	// Asynchronously publish
	for i := 0; i < workers; i++ {
		wg1.Add(1)
		go func() {
			defer wg1.Done()
			for i := 0; i < amount; i++ {
				ebus.Emit(b, MyIntEvent1{})
			}
		}()
	}

	defer ebus.On(b, func(ev MyIntEvent1) {
		// Subscriber that does nothing
	})()

	wg1.Wait()
	wg2.Wait()
	assert.Equal(t, int64(amount*workers), atomic.LoadInt64(&count))
}

type MyChannelEventType int

func (m MyChannelEventType) Hash() uintptr                { return uintptr(m) }
func (m MyChannelEventType) Equals(b ebus.EventType) bool { return m == b }

type MyChannelEvent struct{ Channel int }

func (ev MyChannelEvent) Type() ebus.EventType { return MyChannelEventType(ev.Channel) }

func TestChannelMatrix(t *testing.T) {
	const amount = 1000
	for _, subs := range []int{1, 10, 100} {
		for _, channels := range []int{1, 10} {
			expected := subs * channels * amount
			t.Run(fmt.Sprintf("%dx%d", channels, subs), func(t *testing.T) {
				var count atomic.Int64
				var wg sync.WaitGroup
				wg.Add(expected)

				b := ebus.New()
				for i := 0; i < subs; i++ {
					for channel := 0; channel < channels; channel++ {
						defer ebus.SubscribeTo(b, 1, MyChannelEventType(channel), func(ev MyChannelEvent) {
							count.Add(1)
							wg.Done()
						})()
					}
				}

				for n := 0; n < amount; n++ {
					for id := 0; id < channels; id++ {
						go ebus.Emit(b, MyChannelEvent{Channel: id})
					}
				}

				wg.Wait()
				assert.Equal(t, expected, int(count.Load()))
			})
		}
	}
}
