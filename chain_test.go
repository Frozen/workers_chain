package workers_chain_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/frozen/workers_chain"
	"github.com/stretchr/testify/require"
)

type w1 struct {
}

func (w w1) Work(ch Channel) error {
	for value := range ch.Receive() {
		v := value.(int)
		if v == 0 {
			return errors.New("error first worker")
		}
		ch.Send(v * 10)
	}

	return nil
}

type w2 struct {
}

func (w w2) Work(ch Channel) error {
	for value := range ch.Receive() {
		v := value.(int)
		if v == 0 {
			return errors.New("error second worker")
		}
		ch.Send(v + 2)
	}
	return nil
}

func TestRunWithoutError(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	q := New(ctx, 64)

	manager := q.AddWorker(w1{})

	manager.Send(1)
	manager.Interrupt()

	require.Equal(t, 10, <-manager.Receive())

	err := q.Wait()
	require.NoError(t, err)
}

func TestRunWithError(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	q := New(ctx, 64)

	manager := q.AddWorker(w1{})
	manager.Send(0)

	require.Equal(t, nil, <-manager.Receive())

	err := q.Wait()
	require.Error(t, err)
}

func TestSecondWorkerReceiveMessage(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	q := New(ctx, 64)

	manager1 := q.AddWorker(w1{})
	manager1.Send(1)
	manager1.Interrupt()

	manager2 := q.AddWorker(w2{})

	require.Equal(t, 12, <-manager2.Receive())

	err := q.Wait()
	require.NoError(t, err)
}

func TestSecondWorkerInterruptWithoutError(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	q := New(ctx, 64)

	manager1 := q.AddWorker(w1{})
	manager1.Send(1)

	manager2 := q.AddWorker(w2{})

	require.Equal(t, 12, <-manager2.Receive())
	manager2.Interrupt()

	require.NoError(t, q.Wait())
}

func TestSecondWorkerInterruptWithError(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	q := New(ctx, 64)

	manager1 := q.AddWorker(w1{})
	manager1.Send(1)

	manager2 := q.AddWorker(w2{})

	require.Equal(t, 12, <-manager2.Receive())
	manager2.Send(0)

	err := q.Wait()
	//require.Equal(t, "error second worker", err.Error())
	require.EqualError(t, err, "error second worker")
}

type producer struct {
}

func (p producer) Produce(ctx context.Context, channel OutgoingChannel) error {
	for i := 1; i <= 3; i++ {
		channel.Send(i)
	}
	return nil
}

func TestWorkersWithProducer(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	q := New(ctx, 64)

	q.AddWorker(w1{})
	m2 := q.AddWorker(w2{})
	q.AddProducer(producer{})

	require.Equal(t, 12, <-m2.Receive())
	require.Equal(t, 22, <-m2.Receive())
	require.Equal(t, 32, <-m2.Receive())

	require.NoError(t, q.Wait())
}

func TestNoWorkersWait(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())
	q := New(ctx, 64)

	require.NoError(t, q.Wait())
}

func TestRunWithCancelledContextWithoutWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	q := New(ctx, 64)
	require.NoError(t, q.Wait())
}

func TestRunWithCancelledContextWithWorkers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	q := New(ctx, 64)
	q.AddWorker(w1{})
	require.NoError(t, q.Wait())
}
