package workers_chain

import "sync/atomic"

type IncomingChannel interface {
	Receive() <-chan interface{}
}

type OutgoingChannel interface {
	Send(value interface{})
}

type Channel interface {
	IncomingChannel
	OutgoingChannel
}

type closableChannel interface {
	IncomingChannel
	OutgoingChannel
	Close()
}

type multiChannelImpl struct {
	incoming IncomingChannel
	outgoing closableChannel
}

func (a *multiChannelImpl) Receive() <-chan interface{} {
	return a.incoming.Receive()
}

func (a *multiChannelImpl) Send(value interface{}) {
	a.outgoing.Send(value)
}

func (a *multiChannelImpl) Close() {
	a.outgoing.Close()
}

func newMultiChannel(incoming IncomingChannel, outgoing closableChannel) *multiChannelImpl {
	return &multiChannelImpl{
		incoming: incoming,
		outgoing: outgoing,
	}
}

type channelImpl struct {
	ch     chan interface{}
	closed chan struct{}
	flag   uint32
}

func (a *channelImpl) Send(value interface{}) {
	select {
	case <-a.closed:
	case a.ch <- value:
	}
}

func (a *channelImpl) Receive() <-chan interface{} {
	return a.ch
}

func (a *channelImpl) Close() {
	if atomic.CompareAndSwapUint32(&a.flag, 0, 1) {
		close(a.closed)
		close(a.ch)
	}
}

func newChannel(ch chan interface{}) *channelImpl {
	return &channelImpl{
		ch:     ch,
		closed: make(chan struct{}),
		flag:   0,
	}
}
