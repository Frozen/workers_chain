package workers_chain

import (
	"context"
	"sync"
)

type worker struct {
	w               Worker
	incomingChannel Channel
	outgoingChannel Channel
}

type defaultProducer struct {
}

func (a defaultProducer) Produce(ctx context.Context, _ OutgoingChannel) error {
	<-ctx.Done()
	return nil
}

type chainImpl struct {
	ctx    context.Context
	cancel context.CancelFunc

	// goroutine exit ch
	goroutineExitCh chan struct{}

	// producer output ch
	producerOutputCh Channel

	// size of Channel
	size    uint32
	workers []worker
	mu      sync.Mutex
	// running workers count
	running int

	// first exited goroutine error message(if any)
	err chan error
}

func (a *chainImpl) Wait() error {
	a.mu.Lock()
	workersCnt := len(a.workers)
	a.mu.Unlock()
	if workersCnt == 0 {
		a.cancel()
		return <-a.err
	}

	for {
		<-a.goroutineExitCh
		a.cancel()
		a.mu.Lock()
		a.running -= 1
		running := a.running
		a.mu.Unlock()
		if running == 0 {
			return <-a.err
		}
	}
}

func (a *chainImpl) AddProducer(p Producer) OutgoingChannel {
	a.mu.Lock()
	defer a.mu.Unlock()

	workerChan := newMultiChannel(NewChannel(1), a.producerOutputCh)

	go func() {
		defer func() {
			a.goroutineExitCh <- struct{}{}
		}()
		defer workerChan.Close()
		select {
		case a.err <- p.Produce(a.ctx, workerChan):
		default:
		}
	}()
	return workerChan
}

func (a *chainImpl) AddWorker(w Worker) Manager {
	a.mu.Lock()
	defer a.mu.Unlock()
	var incoming Channel
	if len(a.workers) == 0 {
		incoming = a.producerOutputCh
	} else {
		incoming = a.workers[len(a.workers)-1].outgoingChannel
	}
	outgoing := NewChannel(a.size)
	wk := worker{
		w:               w,
		outgoingChannel: outgoing,
		incomingChannel: incoming,
	}

	a.workers = append(a.workers, wk)
	workerChan := newMultiChannel(incoming, outgoing)
	a.running += 1
	go func() {
		defer func() {
			a.goroutineExitCh <- struct{}{}
		}()
		defer workerChan.Close()

		select {
		case a.err <- w.Work(newMultiChannel(incoming, outgoing)):
		default:
		}
	}()
	return newManager(wk)
}

func New(ctx context.Context, size uint32) Chain {
	ctx, cancel := context.WithCancel(ctx)
	q := &chainImpl{
		size:             size,
		ctx:              ctx,
		cancel:           cancel,
		goroutineExitCh:  make(chan struct{}, 1),
		err:              make(chan error, 1),
		producerOutputCh: NewChannel(size),
	}
	q.AddProducer(defaultProducer{})
	return q
}
