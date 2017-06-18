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
	Branches []Branch
	BranchDictionary map[ID]int
	LastEventID ID
	Reloader Reloader
	StoreCreator StoreCreator
	Attributes interface{}
}

func (t *TimelineStore) GetBranch(id ID) (*Branch, error) {
	branchIndex, ok := t.BranchDictionary[id]
	if !ok {
		return nil, fmt.Errorf("Unknown branch")
	}
	if len(t.Branches) <= branchIndex {
		return nil, fmt.Errorf("Branchindex %d doesn't exist", branchIndex)
	}
	return &t.Branches[branchIndex], nil
}

// NewTimelineStore TODO
func NewTimelineStore(storeCreator StoreCreator, reloader Reloader, attributes interface{}) *TimelineStore {
	gob.Register(storeCreator())
	if attributes != nil {
		gob.Register(attributes)
	}
	return &TimelineStore{
		BranchDictionary: make(map[ID]int),
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
	RewindTime uint64
}

// Type TODO
func (e *NewBranchEvent) Type() EventType {
    return "new_branch"
}

// NewBranch TODO
func NewBranch(time uint64, prevBranch ID, prevBranchLastEvent ID, currentTime uint64, id uint64) *NewBranchEvent {
	randomUUID, _ := uuid.NewRandom()
	return &NewBranchEvent{
		PrevBranch: prevBranch, 
		PrevBranchLastEvent: prevBranchLastEvent,
		NewBranchID: ID(randomUUID),
		BaseEvent: BaseEvent{
			EventID: GenerateTimeID(currentTime, id),
			EventTime: currentTime,
		},
		RewindTime: time,
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
		CreationTime: event.RewindTime,
		StoreID: uint(len(store.Stores)),
		PrevBranch: event.PrevBranch,
		PrevBranchLastEvent: event.PrevBranchLastEvent,
	}
	store.BranchDictionary[event.NewBranchID] = len(store.Branches)
	store.Branches = append(store.Branches, newBranch)
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
		prevBranchIndex, ok := store.BranchDictionary[prevBranch.PrevBranch]
		if !ok {
			break
		}
		if len(store.Branches) <= prevBranchIndex {
			panic(fmt.Errorf("Branchindex %d doesn't exist", prevBranchIndex))
		}
		prevBranch = store.Branches[prevBranchIndex]

		branches = append(branches, branchWithEnd{
			&prevBranch,
			prevBranchLastEvent,
		})
	}
	if len(branches) == 0 {
		return
	}
	lastEventTime := event.RewindTime
	branchIndex := 0
	err := store.Reloader.EventStore.Restore(lastEventTime, func(e Event) error {
		if branchIndex >= len(branches) {
			return nil
		}
		event, ok := e.(TimelineEvent)
		if !ok {
			return nil
		}
		if event.BranchTime() > lastEventTime {
			return nil
		}
		if !EventForBranch(store, &newBranch, event) {
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
		store.Stores[newBranch.StoreID].LastEvent = e
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
	fmt.Println(fmt.Sprintf("Winding back time to %d", event.WindBackTime))
	lastEventTime := event.WindBackTime
	err := store.Reloader.EventStore.Restore(lastEventTime, func(e Event) error {
		event, ok := e.(TimelineEvent)
		if !ok {
			return nil
		}
		fmt.Println(fmt.Sprintf("Rewinding event with time %d and type %s", event.BranchTime(), e.Type()))
		handler, ok := d.TimelineRoutes[e.Type()]
		if !ok {
			return fmt.Errorf("No handler for event")
		}
		if event.BranchTime() > lastEventTime {
			return nil
		}		
		for _, v := range store.Branches {
			if !EventForBranch(store, &v, event) {
				continue
			}
			handler(event, &store.RewindStores[v.StoreID])
			store.RewindStores[v.StoreID].LastEvent = e
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
	LastEventTime uint64
    BranchID ID 
    StoreID uint
	PrevBranch ID
	PrevBranchLastEvent ID
}

// BaseTimelineEvent TODO
type BaseTimelineEvent struct {
	BaseEvent
	BranchID ID
	BranchEventTime uint64
}

// Branch TODO
func (e *BaseTimelineEvent) Branch() ID {
	return e.BranchID
}

func (e *BaseTimelineEvent) BranchTime() uint64 {
	return e.BranchEventTime
}

// TimelineEvent TODO
type TimelineEvent interface {
	Event
	Branch() ID
	BranchTime() uint64
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
	branchIndex, ok := store.BranchDictionary[event.Branch()]
	if !ok {
		panic(fmt.Errorf("Unknown branch: %s, eventtype: %s", event.Branch().ToString(), event.Type()))
	}
	if len(store.Branches) <= branchIndex {
		panic(fmt.Errorf("Branchindex %d doesn't exist", branchIndex))
	}
	branch := store.Branches[branchIndex] 
	handler, ok := d.TimelineRoutes[e.Type()]
	if !ok {
		panic(fmt.Errorf("Unknown Event"))
	}
	branchStore := &store.Stores[branch.StoreID]
	handler(e, branchStore)
	branchStore.LastEvent = e
	eventTime := event.BranchTime()
	if branch.LastEventTime > eventTime {
		return
	}
	fmt.Println(fmt.Sprintf("The new time of the branch is %d", eventTime))
	branch.LastEventTime = eventTime
	store.Branches[branchIndex] = branch
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

func EventForBranch(store *TimelineStore, branch *Branch, event TimelineEvent) bool {
	var recFunc func(branch *Branch, lastEventID ID) bool
	recFunc = func(branch *Branch, lastEventID ID) bool {
		if branch.BranchID != event.Branch() {
			if branch.PrevBranch == ZeroID() {
				return false
			}
			prevBranchIndex, _ := store.BranchDictionary[branch.PrevBranch]
			prevBranch := store.Branches[prevBranchIndex]
			return recFunc(&prevBranch, branch.PrevBranchLastEvent)
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