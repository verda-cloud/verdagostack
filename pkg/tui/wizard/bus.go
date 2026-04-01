package wizard

import "reflect"

type regionSlot struct {
	id      string
	region  Region
	subs    map[reflect.Type]bool
	last    string // current render output
	printed string // last output sent to terminal
}

// MessageBus routes messages between the engine and regions.
type MessageBus struct {
	slots []regionSlot
}

// NewMessageBus creates an empty message bus.
func NewMessageBus() *MessageBus {
	return &MessageBus{}
}

// Register adds a region to the bus.
func (b *MessageBus) Register(id string, r Region) {
	subs := make(map[reflect.Type]bool)
	for _, t := range r.Subscribe() {
		subs[t] = true
	}
	b.slots = append(b.slots, regionSlot{
		id:     id,
		region: r,
		subs:   subs,
	})
}

// Broadcast sends a message to ALL regions (engine-level events).
// Processes any published messages from regions (chained delivery).
func (b *MessageBus) Broadcast(msg any) {
	var pending []any
	for i := range b.slots {
		render, published := b.slots[i].region.Update(msg)
		b.slots[i].last = render
		pending = append(pending, published...)
	}
	b.deliverPending(pending)
}

// Publish sends messages only to regions that subscribed to those types.
// Processes any published messages from regions (chained delivery).
func (b *MessageBus) Publish(from string, msgs []any) {
	var pending []any
	for _, msg := range msgs {
		msgType := reflect.TypeOf(msg)
		for i := range b.slots {
			if b.slots[i].id == from {
				continue
			}
			if b.slots[i].subs[msgType] {
				render, published := b.slots[i].region.Update(msg)
				b.slots[i].last = render
				pending = append(pending, published...)
			}
		}
	}
	b.deliverPending(pending)
}

// deliverPending processes chained messages (published by regions during Update).
func (b *MessageBus) deliverPending(msgs []any) {
	for len(msgs) > 0 {
		var next []any
		for _, msg := range msgs {
			msgType := reflect.TypeOf(msg)
			for i := range b.slots {
				if b.slots[i].subs[msgType] {
					render, published := b.slots[i].region.Update(msg)
					b.slots[i].last = render
					next = append(next, published...)
				}
			}
		}
		msgs = next
	}
}

// RenderAll returns the last rendered output of each region in order.
func (b *MessageBus) RenderAll() []string {
	renders := make([]string, len(b.slots))
	for i, s := range b.slots {
		renders[i] = s.last
	}
	return renders
}

// RenderChanged returns outputs only for regions whose render changed
// since the last call to RenderChanged. Unchanged regions return "".
func (b *MessageBus) RenderChanged() []string {
	renders := make([]string, len(b.slots))
	for i := range b.slots {
		if b.slots[i].last != b.slots[i].printed {
			renders[i] = b.slots[i].last
			b.slots[i].printed = b.slots[i].last
		}
	}
	return renders
}
