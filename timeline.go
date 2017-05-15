package event

import (
	"bytes"
	"fmt"
	"encoding/gob"

	"github.com/google/uuid"
)

type StoreCreator func() BranchStore

// TimelineStore TODO
type TimelineStore struct {
	Stores []Store
	RewindStores []Store
	Branches map[ID]Branch
	LastEventID ID
	Reloader Reloader
	StoreCreator StoreCreator
	Attributes interface{}
}

// NewTimelineStore TODO
func NewTimelineStore(storeCreator StoreCreator, reloader Reloader, attributes interface{}) *TimelineStore {
	gob.Register(storeCreator())
	if attributes != nil {
		gob.Register(attributes)
	}
	return &TimelineStore{
		Branches: make(map[ID]Branch),
		LastEventID: GenerateTimeID(0,0),
		StoreCreator: storeCreator,
		Reloader: reloader,
		Attributes: attributes,
	}
}

// NewBranchEvent TODO
type NewBranchEvent struct {
	BaseEvent
	PrevBranch ID
	PrevBranchLastEvent ID
	NewBranchID ID
}

// Type TODO
func (e *NewBranchEvent) Type() EventType {
    return "new_branch"
}

// NewBranch TODO
func NewBranch(prevBranch ID, prevBranchLastEvent ID, currentTime uint64, id uint64) *NewBranchEvent {
	randomUUID, _ := uuid.NewRandom()
	return &NewBranchEvent{
		PrevBranch: prevBranch, 
		PrevBranchLastEvent: prevBranchLastEvent,
		NewBranchID: ID(randomUUID),
		BaseEvent: BaseEvent{
			EventID: GenerateTimeID(currentTime, id),
			EventTime: currentTime,
		},
	}
}

type branchWithEnd struct{
	Branch *Branch
	LastEvent ID
}

// NewBranchHandler TODO
func (d* TimelineDispatcher) NewBranchHandler(e Event, s *Store) {
	event, ok := e.(*NewBranchEvent)
	if !ok {
		panic(fmt.Errorf("Event not of right type"))
	}
	store, ok := s.Attributes.(*TimelineStore)
	if !ok {
		panic(fmt.Errorf("Event not of right type"))
	}
	newBranch := Branch {
		BranchID: event.NewBranchID,
		CreationTime: e.Time(),
		StoreID: uint(len(store.Stores)),
		PrevBranch: event.PrevBranch,
		PrevBranchLastEvent: event.PrevBranchLastEvent,
	}
	store.Branches[event.NewBranchID] = newBranch
	store.Stores = append(store.Stores, Store{
		Attributes: store.StoreCreator(),
	})
	store.RewindStores = append(store.RewindStores, Store{
		Attributes: store.StoreCreator(),
	})

	branches := []branchWithEnd{}
	prevBranch := newBranch
	for {
		if ZeroID() == prevBranch.PrevBranch{
			break
		}
		prevBranchLastEvent := prevBranch.PrevBranchLastEvent
		prevBranch, ok = store.Branches[prevBranch.PrevBranch]
		if !ok {
			break
		}

		branches = append(branches, branchWithEnd{
			&prevBranch,
			prevBranchLastEvent,
		})
	}
	if len(branches) == 0 {
		return
	}

	branchIndex := 0
	err := store.Reloader.EventStore.Restore(^uint64(0), func(e Event) error {
		if branchIndex >= len(branches) {
			return nil
		}
		event, ok := e.(TimelineEvent)
		if !ok {
			return nil
		}
		if event.Branch() != branches[branchIndex].Branch.BranchID {
			return nil
		}
		if event.ID() == branches[branchIndex].LastEvent {
			branchIndex++
		}
		handler, ok := d.TimelineRoutes[e.Type()]
		if !ok {
			return fmt.Errorf("No handler for event")
		}
		handler(event, &store.Stores[newBranch.StoreID])
		return nil
	})
	if nil != err {
		panic(err)
	}
}

// WindbackEvent TODO
type WindbackEvent struct {
	BaseEvent
	WindBackTime uint64
}

// Type TODO
func (e *WindbackEvent) Type() EventType {
    return "windback"
}

// Windback TODO
func Windback(time uint64, currentTime uint64, id uint64) *WindbackEvent {
	return &WindbackEvent{
		WindBackTime: time,
		BaseEvent: BaseEvent{
			EventID: GenerateTimeID(currentTime, id),
			EventTime: currentTime,
		},
	}
}

// WindbackHandler TODO
func (d* TimelineDispatcher) WindbackHandler(e Event, s *Store) {
	event, ok := e.(*WindbackEvent)
	if !ok {
		panic(fmt.Errorf("Event not of right type"))
	}
	store, ok := s.Attributes.(*TimelineStore)
	if !ok {
		panic(fmt.Errorf("Event not of right type"))
	}
	for k, v := range store.Stores {
		branchID := v.Attributes.(BranchStore).GetBranchID()
		freshStore := store.StoreCreator()
		freshStore.SetBranchID(branchID)
		store.RewindStores[k] = Store {
			Attributes: freshStore,
		}
	}

	err := store.Reloader.EventStore.Restore(event.WindBackTime, func(e Event) error {
		//fmt.Println("test1")
		event, ok := e.(TimelineEvent)
		if !ok {
			return nil
		}
		handler, ok := d.TimelineRoutes[e.Type()]
		if !ok {
			return fmt.Errorf("No handler for event")
		}
		
		for _, v := range store.Branches {
			//fmt.Println("test2")
			if EventForBranch(store.Branches, &v, event) {
				fmt.Println(v.BranchID)
				handler(event, &store.RewindStores[v.StoreID])
			}	
		}
		return nil
	})
	//TODO return error
	if nil != err {
		panic(err)
	}
}

// Branch TODO
type Branch struct {
    CreationTime uint64 
    BranchID ID 
    StoreID uint
	PrevBranch ID
	PrevBranchLastEvent ID
}

// BaseTimelineEvent TODO
type BaseTimelineEvent struct {
	BaseEvent
	BranchID ID
}

// Branch TODO
func (e *BaseTimelineEvent) Branch() ID {
	return e.BranchID
}

// TimelineEvent TODO
type TimelineEvent interface {
	Event
	Branch() ID
}

// TimelineDispatcher TODO
type TimelineDispatcher struct {
	Dispatcher
	TimelineRoutes map[EventType] HandlerFunc
}

// NewTimelineDispatcher TODO
func NewTimelineDispatcher(timeStore *TimelineStore) *TimelineDispatcher {
	store := NewStore(timeStore)

	return &TimelineDispatcher{
		Dispatcher: Dispatcher{
			Handlers: make(map[EventType] HandlerFunc),
			Store: store,
		},
		TimelineRoutes: make(map[EventType] HandlerFunc),
	}
}


// TimelineEventHandler TODO
func (d* TimelineDispatcher) TimelineEventHandler(e Event, s *Store) {
	store, ok := s.Attributes.(*TimelineStore)
    if !ok {
        panic(fmt.Errorf("Store not of right type"))
    }
	event := e.(TimelineEvent)
	branch, ok := store.Branches[event.Branch()]
	if !ok {
		panic(fmt.Errorf("Unknown branch"))
	}
	handler, ok := d.TimelineRoutes[e.Type()]
	if !ok {
		panic(fmt.Errorf("Unknown Event"))
	}
	branchStore := &store.Stores[branch.StoreID]
	handler(e, branchStore)
	branchStore.LastEvent = e
}

// Register TODO
func (d* TimelineDispatcher) Register(event Event, handler HandlerFunc) {
	d.TimelineRoutes[event.Type()] = handler
	d.Dispatcher.Register(event, d.TimelineEventHandler)
}

// Reloader TODO
type Reloader struct {
	EventStore *BoltEventStore
	SnapShotStore *BoltSnapshotStore
}

type BranchStore interface {
	GetBranchID() ID
	SetBranchID(id ID)
}

func EventForBranch(branches map[ID]Branch, branch *Branch, event TimelineEvent) bool {
	var recFunc func(branch *Branch, lastEventID ID) bool
	recFunc = func(branch *Branch, lastEventID ID) bool {
		if branch.BranchID != event.Branch() {
			if branch.PrevBranch == ZeroID() {
				return false
			}
			b, _ := branches[branch.PrevBranch]
			return recFunc(&b, branch.PrevBranchLastEvent)
		}
		id1 := lastEventID[:]
		eventID := event.ID()
		id2 := eventID[:]
		
		if bytes.Compare(id1, id2) == -1  {
			return false
		}

		return true
	}
	return recFunc(branch, MaxID())
}