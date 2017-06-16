package event

// EventStoreMiddleware TODO
func EventStoreMiddleware(store EventStore) MiddlewareFunc{
    return func(e Event) {
        store.Add(e)
    }
}

// ReadEventHandleFunc TODO
type ReadEventHandleFunc func(e Event) error

// EventStore TODO
type EventStore interface {
    Add(e Event)
    Restore(time uint64, handleFunc ReadEventHandleFunc) error
}

// EventStoreMem TODO
type EventStoreMem struct {
    Events []Event
}

// Add TODO
func (s *EventStoreMem) Add(e Event) {
    s.Events = append(s.Events, e)
}

// Restore TODO
func (s *EventStoreMem) Restore(time uint64, handleFunc ReadEventHandleFunc) error {
    for _, e := range s.Events {
        handleFunc(e)
    }

    return nil
}

type EventHandler interface {
    Handle(Event) error
}

func RestoreEvents(e EventStore, d EventHandler) {
    e.Restore(^uint64(0), d.Handle)
}