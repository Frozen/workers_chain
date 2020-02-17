package workers_chain

import (
	"testing"
)

func TestChannelMultipleClose(t *testing.T) {
	ch := NewChannel(10)
	ch.Close()
	ch.Close()
}

func TestChannelUnlockOnClose(t *testing.T) {
	ch := NewChannel(1)
	go func() {
		ch.Close()
	}()
	ch.Send(1)
	ch.Send(2)
}
