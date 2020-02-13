package workers_chain

import "context"

type funcWorker struct {
	f func(Channel) error
}

func (a funcWorker) Work(ch Channel) error {
	return a.f(ch)
}

func FuncWorker(f func(Channel) error) Worker {
	return funcWorker{f: f}
}

type funcProducer struct {
	f func(ctx context.Context, ch OutgoingChannel) error
}

func (a funcProducer) Produce(ctx context.Context, channel OutgoingChannel) error {
	return a.f(ctx, channel)
}

func FuncProducer(f func(ctx context.Context, ch OutgoingChannel) error) Producer {
	return funcProducer{f: f}
}
