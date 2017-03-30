package event

func EventStoreMiddleware(store EventStore) MiddlewareFunc{
    return func(e *Event) {
        store.Add(e)
    }
}

type EventStore interface {
    Add(e *Event)
    Restore(d *Dispatcher)
}

type EventStoreMem struct {
    Events []*Event
}

func (s *EventStoreMem) Add(e *Event) {
    s.Events = append(s.Events, e)
}

func (s *EventStoreMem) Restore(d *Dispatcher) {
    for _, e := range s.Events {
        d.Handle(e)
    }
}