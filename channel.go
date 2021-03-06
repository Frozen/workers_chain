package workers_chain

import (
	"sync"
	"sync/atomic"
)

type Channel interface {
	Send(val interface{}) bool
	Receive() (interface{}, bool)
	Close()
}

type OutgoingChannel interface {
	Send(val interface{}) bool
	Close()
}

type ChannelImpl struct {
	cond   *sync.Cond
	values []interface{}
	index  uint32
	// count of messages in queue
	len uint32
	// size(capacity) of queue
	size uint32
	// closed
	closed uint32
}

func NewChannel(size uint32) *ChannelImpl {
	return &ChannelImpl{
		cond:   sync.NewCond(&sync.Mutex{}),
		values: make([]interface{}, size),
		size:   size,
	}
}
func (a *ChannelImpl) Send(val interface{}) bool {
	a.cond.L.Lock()
	for {
		if 1 == atomic.LoadUint32(&a.closed) {
			return false
		}
		// full of capacity
		if a.len == a.size {
			a.cond.Wait()
			continue
		}
		index := (a.index + a.len) % a.size
		a.values[index] = val
		a.len++
		break
	}
	a.cond.Signal()
	a.cond.L.Unlock()
	return true
}
func (a *ChannelImpl) Receive() (interface{}, bool) {
	a.cond.L.Lock()
	defer a.cond.L.Unlock()
	for {
		if a.len > 0 {
			val := a.values[a.index]
			a.index = (a.index + 1) % a.size
			a.len--
			a.cond.Broadcast()
			return val, true
		}
		if 1 == atomic.LoadUint32(&a.closed) {
			return nil, false
		}
		a.cond.Wait()
	}
}

func (a *ChannelImpl) Close() {
	atomic.StoreUint32(&a.closed, 1)
	a.cond.Broadcast()
}

type multiChannel struct {
	incoming Channel
	outgoing Channel
}

func (a *multiChannel) Send(val interface{}) bool {
	return a.outgoing.Send(val)
}

func (a *multiChannel) Receive() (interface{}, bool) {
	return a.incoming.Receive()
}

func (a *multiChannel) Close() {
	a.incoming.Close()
	a.outgoing.Close()
}

func newMultiChannel(incoming, outgoing Channel) Channel {
	return &multiChannel{
		incoming: incoming,
		outgoing: outgoing,
	}
}
