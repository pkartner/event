package main

import (
    "fmt"

    "github.com/boltdb/bolt"

    "github.com/pkartner/event"
)

type Store struct {
    Counter int
}

type IncreaseCounterEvent struct {
    Amount int
}

type DecreaseCounterEvent struct {
    Amount int
}

func GetInnerStore(s *event.Store) *Store {
    store, ok := s.Attributes.(*Store)
    if !ok {
        panic(fmt.Errorf("Store not of right type"))
    }
    return store
} 

func IncreaseCounter(timeline event.ID, amount int) *event.Event {
    return &event.Event{
        Attributes: &IncreaseCounterEvent{
            Amount: amount,
        },
        Type: "increase_counter",
        ID: "something",
        Time: 0,
        TimeLineID: timeline,
    }
}

func DecreaseCounter(timeline event.ID, amount int) *event.Event {
    return &event.Event{
        Attributes: &DecreaseCounterEvent{
            Amount: amount,
        },
        Type: "decrease_counter",
        ID: "something",
        Time: 0,
        TimeLineID: timeline,
    }
}

func IncreaseCounterHandler(e *event.Event, s *event.Store) {
    event, ok := e.Attributes.(*IncreaseCounterEvent)
    if !ok {
        panic(fmt.Errorf("Event not of right type"))
    }
    store := GetInnerStore(s)

    store.Counter += event.Amount
}

func DecreaseCounterHandler(e *event.Event, s *event.Store) {
    event, ok := e.Attributes.(*DecreaseCounterEvent)
    if !ok {
        panic(fmt.Errorf("Event not of right type"))
    }
    store := GetInnerStore(s)

    store.Counter -= event.Amount
}

func Dispatch(d *event.Dispatcher, e *event.Event) {
    if err := d.Dispatch(e); nil != err {
        panic(err)
    }
}

func main() {
    timeIDGenerator := &event.TimeIDGenerator{}
    innerStore := &Store{}
    store := event.NewStore("0", innerStore)
    db, err := bolt.Open("event.db", 0600, nil)
    if nil != err {
        panic(err)
    }
    defer db.Close()
    stateStore := event.NewBoltSnapshotStore(db)
    eventStore := event.NewBoltEventStore(db)
    dispatcher := event.NewDispatcher()
    dispatcher.AddStore(store)
    dispatcher.SetMiddleware(
        event.EventStoreMiddleware(eventStore),
    )
    dispatcher.SetHandler("increase_counter", &IncreaseCounterEvent{}, IncreaseCounterHandler)
    dispatcher.SetHandler("decrease_counter", &DecreaseCounterEvent{}, DecreaseCounterHandler)

    //eventStore.Restore(dispatcher)

    // Dispatch(dispatcher, IncreaseCounter(store.ID, 2))
    // Dispatch(dispatcher, DecreaseCounter(store.ID, 3))
    // if err := stateStore.Write(store); nil != err {
    //     panic(err)
    // } 
    
    restored, err := stateStore.Restore()
    if nil != err {
        panic(err)
    }
    innerStore = restored.Attributes.(*Store)

    fmt.Println(innerStore.Counter)
}

