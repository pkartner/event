package event

import (
    "fmt"
    "encoding/gob"
)

type ID [16]byte

type Event struct {
    ID ID
    StoreID ID
    Type string
    Time uint64
    Attributes interface{}
}

type Dispatcher struct {
    Middlewares []Middleware
    Stores map[ID]*Store
    Handlers map[string] HandlerFunc
}

func NewDispatcher() *Dispatcher {
    dispatcher := Dispatcher{}
    dispatcher.Stores = map[ID]*Store{}
    dispatcher.Handlers = map[string]HandlerFunc{}
    return &dispatcher
}

func (d *Dispatcher) AddStore(s *Store) {
    d.Stores[s.ID] = s
}

func (d *Dispatcher) Dispatch(event *Event) error{
    for _, m :=  range d.Middlewares {
        m.Do(event)
    }
    return d.Handle(event)
}

func (d *Dispatcher) SetHandler(eventType string, event interface{}, handler HandlerFunc) {
    gob.Register(event)
    d.Handlers[eventType] = handler
}

// Handle should not be called externally, use Dispatch instead, this function gives the possibility to avoid the middleware
func (d *Dispatcher) Handle(event *Event) error{
    store, ok := d.Stores[event.BranchID]
    if !ok {
        return fmt.Errorf("Timeline with ID: %s is unknown", event.BranchID)
    }
    handler, ok := d.Handlers[event.Type]
    if !ok {
        return fmt.Errorf("Event with type: %s has no handler", event.Type)
    }
    handler(event, store)
    store.Time = event.Time
    return nil
}

func (d *Dispatcher) SetMiddleware(m ...Middleware) {
    d.Middlewares = m
}

type Middleware interface{
    Do(event *Event)
}
type MiddlewareFunc func(e *Event)
func (m MiddlewareFunc) Do(e *Event) {
    m(e)
}

type HandlerFunc func(event *Event, store *Store)

type Store struct {
    ID ID
    Time uint64
    Branch *Branch
    Attributes interface{}
}

// NewStore accepts your attribute stores and wraps it in an event.Store object
func NewStore(id ID, store interface{}) *Store {
    gob.Register(store)
    return &Store {
        Attributes: store,
    }
}

