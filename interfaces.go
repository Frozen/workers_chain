package workers_chain

import "context"

type Producer interface {
	Produce(ctx context.Context, channel OutgoingChannel) error
}

type Worker interface {
	Work(Channel) error
}

type Chain interface {
	AddWorker(Worker) Manager
	AddProducer(Producer) OutgoingChannel
	Wait() error
}
