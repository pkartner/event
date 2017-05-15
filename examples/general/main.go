package main

import (
    "fmt"
    "os"
    "time"

    "github.com/boltdb/bolt"

    "github.com/pkartner/event"
)

// Store TODO
type Store struct {
    BranchID event.ID
    Counter int
}

// NewStore TODO
func NewStore() event.BranchStore {
    return &Store{}
}

// GetBranchID TODO
func(s *Store) GetBranchID() event.ID{
    return s.BranchID
}

// SetBranchID TODO
func(s *Store) SetBranchID(id event.ID) {
    s.BranchID = id
}

// IncreaseCounterEvent TODO
type IncreaseCounterEvent struct {
    event.BaseTimelineEvent
    Amount int
}

// Type TODO
func (e *IncreaseCounterEvent) Type() event.EventType {
    return "increase_counter"
}

// DecreaseCounterEvent TODO
type DecreaseCounterEvent struct {
    event.BaseTimelineEvent
    Amount int
}

// Type TODO
func (e *DecreaseCounterEvent) Type() event.EventType {
    return "decrease_counter"
}

// GetInnerStore TODO
func GetInnerStore(s *event.Store) *Store {
    store, ok := s.Attributes.(*Store)
    if !ok {
        panic(fmt.Errorf("Store not of right type"))
    }
    return store
} 

// IncreaseCounter TODO
func IncreaseCounter(amount int, currentTime uint64, id uint64, branchID event.ID) event.Event {
    e := IncreaseCounterEvent {}
	e.EventID = event.GenerateTimeID(currentTime, id)
	e.EventTime = currentTime
    e.BranchID = branchID
    e.Amount = amount
    return &e
}

// DecreaseCounter TODO
func DecreaseCounter(amount int, currentTime uint64, id uint64, branchID event.ID) event.Event {
    e := DecreaseCounterEvent {}
	e.EventID = event.GenerateTimeID(currentTime, id)
	e.EventTime = currentTime
    e.BranchID = branchID
    e.Amount = amount
    return &e
}

// IncreaseCounterHandler TODO
func IncreaseCounterHandler(e event.Event, s *event.Store) {
    event, ok := e.(*IncreaseCounterEvent)
    if !ok {
        panic(fmt.Errorf("Event not of right type"))
    }
    store := GetInnerStore(s)
    if nil == store {
        panic(fmt.Errorf("Store is nil"))
    }
    store.Counter += event.Amount
}

// DecreaseCounterHandler TODO
func DecreaseCounterHandler(e event.Event, s *event.Store) {
    event, ok := e.(*DecreaseCounterEvent)
    if !ok {
        panic(fmt.Errorf("Event not of right type"))
    }
    store := GetInnerStore(s)

    store.Counter -= event.Amount
}

// Dispatch TODO
func Dispatch(
    d interface {
        Dispatch(e event.Event) error
    }, e event.Event) {
    if err := d.Dispatch(e); nil != err {
        panic(err)
    }
}

func main() {
    err := os.Remove("event.db")
    if nil != err {
        panic(err)
    }
    db, err := bolt.Open("event.db", 0600, nil)
    if nil != err {
        panic(err)
    }
    defer db.Close()
    //stateStore := event.NewBoltSnapshotStore(db)
    eventStore := event.NewBoltEventStore(db)
    store := event.NewTimelineStore(NewStore, event.Reloader{
        EventStore: eventStore,
    })
    dispatcher := event.NewTimelineDispatcher(store)
    dispatcher.SetMiddleware(
        event.EventStoreMiddleware(eventStore),
    )
    dispatcher.Dispatcher.Register(&event.WindbackEvent{}, dispatcher.WindbackHandler)
    dispatcher.Dispatcher.Register(&event.NewBranchEvent{}, dispatcher.NewBranchHandler)
    dispatcher.Register(&IncreaseCounterEvent{}, IncreaseCounterHandler)
    dispatcher.Register(&DecreaseCounterEvent{}, DecreaseCounterHandler)

    var nextEventID uint64 = 1
    getNextEventID := func() uint64{
        returnValue := nextEventID
        nextEventID++
        return returnValue
    }
    branchEvent := event.NewBranch(event.ZeroID(), event.ZeroID(), 0, getNextEventID())
    branchIDOne := branchEvent.NewBranchID
    //eventStore.Restore(dispatcher)
    Dispatch(dispatcher, branchEvent)
    Dispatch(dispatcher, IncreaseCounter(1, 0, getNextEventID(), branchIDOne))
    branchOneLastEvent := IncreaseCounter(2, 0, getNextEventID(), branchIDOne)
    Dispatch(dispatcher, branchOneLastEvent)
    //Dispatch(dispatcher, DecreaseCounter(3, 1, getNextEventID(), branchIDOne))
    //time.Sleep(time.Second*2)
    //Dispatch(dispatcher, event.Windback(1, 2, getNextEventID()))
    secondBranchEvent := event.NewBranch(branchIDOne, branchOneLastEvent.ID(), 2, getNextEventID())
    branchIDTwo := secondBranchEvent.NewBranchID
    Dispatch(dispatcher, secondBranchEvent)

    // if err := stateStore.Write(store); nil != err {
    //     panic(err)
    // } 
    
    // restored, err := stateStore.Restore()
    // if nil != err {
    //     panic(err)
    // }
    // innerStore = restored.Attributes.(*Store)
    storeID := store.Branches[branchIDOne].StoreID
    innerStoreOne := store.Stores[storeID]
    storeID = store.Branches[branchIDTwo].StoreID
    innerStoreTwo := store.Stores[storeID]
    fmt.Println("After new branch created time 1")
    fmt.Println(innerStoreOne.Attributes.(*Store).Counter)
    fmt.Println(innerStoreTwo.Attributes.(*Store).Counter)

    Dispatch(dispatcher, IncreaseCounter(3, 2, getNextEventID(), branchIDOne))
    Dispatch(dispatcher, IncreaseCounter(4, 2, getNextEventID(), branchIDTwo))
    fmt.Println("After two more events time 2")
    fmt.Println(innerStoreOne.Attributes.(*Store).Counter)
    fmt.Println(innerStoreTwo.Attributes.(*Store).Counter)
    time.Sleep(time.Second)

    eventStore.Restore(^uint64(0), func(e event.Event) error {
        fmt.Println(e.ID())
        return nil
    })

    Dispatch(dispatcher, event.Windback(2, 3, getNextEventID()))
    storeID = store.Branches[branchIDOne].StoreID
    innerStoreOne = store.Stores[storeID]
    storeID = store.Branches[branchIDTwo].StoreID
    innerStoreTwo = store.Stores[storeID]
    fmt.Println("After rewind time 2")
    fmt.Println(innerStoreOne.Attributes.(*Store).Counter)
    fmt.Println(innerStoreTwo.Attributes.(*Store).Counter)
}

