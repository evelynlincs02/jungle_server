package event

type Message interface{}

type EventEmitter map[string][]chan Message

const (
	STOP = "STOP"
)

func (emitter EventEmitter) On(evt string, handler func(args Message)) {
	newC := make(chan Message)
	emitter[evt] = append(emitter[evt], newC)
	go func() {
		for {
			select {
			case msg := <-newC:
				if msg == STOP { // 不確定回收newC之後goroutine會不會自動停止，放心安的
					return
				}
				handler(msg)
			}
		}
	}()
}

func (emitter EventEmitter) RemoveListener(evt string) {
	emitter.Emit(evt, STOP) // to stop the goroutine of the channel receiver
	delete(emitter, evt)
}

func (emitter EventEmitter) RemoveAllListener() {
	for k := range emitter {
		emitter.Emit(k, STOP)
		delete(emitter, k)
	}
}

func (emitter EventEmitter) Emit(evt string, msg Message) {
	if _, ok := emitter[evt]; ok {
		for _, ch := range emitter[evt] {
			// 若不開 goroutine，在On之前就先Emit了的話會deadlock
			go func(channel chan Message) {
				channel <- msg
				if msg == STOP {
					return
				}
			}(ch)
		}
	}
}
