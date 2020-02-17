package workers_chain

type managerImpl struct {
	w worker
}

func (a *managerImpl) Receive() (interface{}, bool) {
	return a.w.outgoingChannel.Receive()
}

func (a *managerImpl) Send(value interface{}) bool {
	return a.w.incomingChannel.Send(value)
}

func (a *managerImpl) Interrupt() {
	a.w.incomingChannel.Close()
}

func newManager(w worker) *managerImpl {
	return &managerImpl{w: w}
}
