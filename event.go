package event

import (
    "fmt"
    "encoding/gob"
)

// ID TODO
type ID [16]byte

func (id ID) Byte() []byte {
    return id[:]
}

// ZeroID TODO
func ZeroID() ID {
    return [16]byte{}
}
func MaxID() ID {
    return GenerateTimeID(^uint64(0), ^uint64(0))
}

// EventType TODO
type EventType string

func (e EventType) String() string {
    return string(e)
}

// Event TODO
type Event interface {
    ID() ID
    Type() EventType
    Time() uint64
}

// BaseEvent TODO
type BaseEvent struct {
    EventID ID
    EventTime uint64
}

// ID TODO
func (e *BaseEvent) ID() ID {
    return e.EventID
}

// Time TODO
func (e *BaseEvent) Time() uint64 {
    return e.EventTime
}

// Dispatcher TODO
type Dispatcher struct {
    Middlewares []Middleware
    Store *Store
    Handlers map[EventType] HandlerFunc
}

// NewDispatcher TODO
func NewDispatcher(s *Store) *Dispatcher {
    dispatcher := Dispatcher{
        Store: s,
    }
    dispatcher.Handlers = map[EventType]HandlerFunc{}
    return &dispatcher
}

// Dispatch TODO
func (d *Dispatcher) Dispatch(event Event) error{
    for _, m :=  range d.Middlewares {
        m.Do(event)
    }
    return d.Handle(event)
}

// Register TODO
func (d *Dispatcher) Register(event Event, handler HandlerFunc) {
    gob.Register(event)
    d.Handlers[event.Type()] = handler
}

// Handle should not be called externally, use Dispatch instead, this function gives the possibility to avoid the middleware
func (d *Dispatcher) Handle(event Event) error{
    handler, ok := d.Handlers[event.Type()]
    if !ok {
        return fmt.Errorf("Event with type: %s has no handler", event.Type())
    }
    handler(event, d.Store)
    d.Store.LastEvent = event
    d.Store.Time = event.Time()
    return nil
}

// SetMiddleware TODO
func (d *Dispatcher) SetMiddleware(m ...Middleware) {
    d.Middlewares = m
}

// Middleware TODO
type Middleware interface{
    Do(event Event)
}
// MiddlewareFunc TODO
type MiddlewareFunc func(e Event)
// Do TODO
func (m MiddlewareFunc) Do(e Event) {
    m(e)
}

// HandlerFunc TODO
type HandlerFunc func(event Event, store *Store)

// Store TODO
type Store struct {
    LastEvent Event
    Time uint64
    Attributes interface{}
}

func (s *Store) NextID() uint64 {
    return s.LastEvent.ID().IDPart()+1
}

// NewStore accepts your attribute stores and wraps it in an event.Store object
func NewStore(store interface{}) *Store {
    gob.Register(store)
    return &Store {
        Attributes: store,
    }
}