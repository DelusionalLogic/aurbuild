package bus

type Handler func(interface{})

type Queue struct {
	handlers []Handler
}

func MakeQueue() Queue {
	return Queue{
		handlers: make([]Handler, 0, 2),
	}
}

func (q *Queue) Subscribe(handler Handler) {
	q.handlers = append(q.handlers, handler)
}

func (q *Queue) Post(payload interface{}) {
	for _, handler := range q.handlers {
		go handler(payload)
	}
}
