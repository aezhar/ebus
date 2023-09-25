// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *
// Copyright(c) 2023 individual contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License
// that can be found in the LICENSE file.
// * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * *

package ebus

import (
	"github.com/aezhar/haxxmap"
)

type Bus struct {
	groups *haxxmap.Map[EventType, any]
}

func New() *Bus {
	groups := haxxmap.New[EventType, any](
		haxxmap.WithHasher[EventType, any](EventType.Hash),
		haxxmap.WithComparator[EventType, any](EventType.Equals),
	)

	return &Bus{
		groups: groups,
	}
}
