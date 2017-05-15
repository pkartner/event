package main

import (
    // "os/exec"
    "fmt"
    "io"
    "log"
    "os"
    "strconv"
    "strings"
    "time"

    "github.com/boltdb/bolt"

    "github.com/chzyer/readline"

    "github.com/pkartner/event"
)

type MainStore struct {
    id uint64
}

// Store TODO
type BranchStore struct {
    BranchID event.ID
    Counter int
}

// NewStore TODO
func NewBranchStore() event.BranchStore {
    return &BranchStore{}
}

// GetBranchID TODO
func(s *BranchStore) GetBranchID() event.ID{
    return s.BranchID
}

// SetBranchID TODO
func(s *BranchStore) SetBranchID(id event.ID) {
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
func GetInnerStore(s *event.Store) *BranchStore {
    store, ok := s.Attributes.(*BranchStore)
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

func AskForBranchID(l *readline.Instance, store *event.TimelineStore) (event.ID, error) {
    tmpAutoCompl := l.Config.AutoComplete
    branches := []readline.PrefixCompleterInterface{}
    for k := range store.Branches {
        branches = append(branches, readline.PcItem(k.ToString()))
    }
    l.Config.AutoComplete = readline.NewPrefixCompleter(branches...)
    defer func() {
        l.Config.AutoComplete = tmpAutoCompl
    }()
    println("What is the branch id?")
    line, err := l.Readline()
    if err == readline.ErrInterrupt {
        return event.ZeroID(), err
    }
    line = strings.TrimSpace(line)
    return event.IDFromString(line), nil
}

func main() {
    // Setup command line stuff 
    l, err := readline.New(">>")
    if nil != err {
        panic(err)
    }
    defer l.Close()
    log.SetOutput(l.Stderr())
    // cmd := exec.Command("ls")
    // cmd.Stderr = l.Stderr()
    // cmd.Stdout = l.Stdout()
    // cmd.Run()

    // Init input
    var databaseFileName string
    newGame := true
    for {
        println("Please fill in a name for you database file!")
        line, err := l.Readline()
        if nil != err {
           panic(err)
        }
        line = strings.TrimSpace(line)
        databaseFileName = line+".db"
        if _, err := os.Stat(databaseFileName); os.IsNotExist(err) {
            break;
        }
        println(databaseFileName, " already exists, do you want to overwrite it? y/n")
        line, err = l.Readline()
        if nil != err {
            panic(err)
        }
        line = strings.TrimSpace(line)
        if line == "y" || line == "yes" {
            if err := os.Remove(databaseFileName); nil != err {
                panic(err)
            }
            break;
        }

        println("Do you want to continue? y/n")
        line, err = l.Readline()
        if nil != err {
            panic(err)
        }
        line = strings.TrimSpace(line)
        if line == "y" || line == "yes" {
            newGame = false
            break;
        }
    }

    // Setup Event stuff
    db, err := bolt.Open(databaseFileName, 0600, nil)
    if nil != err {
        panic(err)
    }
    //defer db.Close()
    //stateStore := event.NewBoltSnapshotStore(db)
    eventStore := event.NewBoltEventStore(db)
    timeStore := event.NewTimelineStore(NewBranchStore, event.Reloader{
        EventStore: eventStore,
    }, nil)
    dispatcher := event.NewTimelineDispatcher(timeStore)
    dispatcher.SetMiddleware(
        event.EventStoreMiddleware(eventStore),
    )
    store := dispatcher.Store
    dispatcher.Dispatcher.Register(&event.WindbackEvent{}, dispatcher.WindbackHandler)
    dispatcher.Dispatcher.Register(&event.NewBranchEvent{}, dispatcher.NewBranchHandler)
    dispatcher.Register(&IncreaseCounterEvent{}, IncreaseCounterHandler)
    dispatcher.Register(&DecreaseCounterEvent{}, DecreaseCounterHandler)

    if newGame {
        branchEvent := event.NewBranch(event.ZeroID(), event.ZeroID(), 0, 0)
        Dispatch(dispatcher, branchEvent)
    } else {
        event.RestoreEvents(eventStore, dispatcher)
    }

    // Main loop
    for {
        line, err := l.Readline()
        if err == readline.ErrInterrupt {
            if len(line) == 0 {
                break
            }
            continue;
        } else if err == io.EOF {
            break
        }
        line = strings.TrimSpace(line)
        switch {
        case strings.HasPrefix(line, "increase"):
            line := strings.TrimSpace(line[8:])
            number, err := strconv.Atoi(line)
            if nil != err {
                log.Println("Can only increase by a number")
                break
            }
            branchID, err := AskForBranchID(l, timeStore)
            if nil != err {
                panic(err)
            }
            Dispatch(dispatcher, IncreaseCounter(number, uint64(time.Now().Unix()), store.NextID(), branchID))
        case strings.HasPrefix(line, "decrease"):
            line := strings.TrimSpace(line[8:])
            number, err := strconv.Atoi(line)
            if nil != err {
                log.Println("Can only increase by a number")
                break
            }
            branchID, err := AskForBranchID(l, timeStore)
            if nil != err {
                panic(err)
            }
            Dispatch(dispatcher, DecreaseCounter(number, uint64(time.Now().Unix()), store.NextID(), branchID))
        case line == "info":
            println("CurrentTime is: ", uint64(time.Now().Unix()))
            for _, v := range timeStore.Branches {
                println("============================")
                innerStoreAbstract := timeStore.Stores[v.StoreID]
                innerStore := innerStoreAbstract.Attributes.(*BranchStore)
                rewindStoreAbstract := timeStore.RewindStores[v.StoreID]
                rewindStore := rewindStoreAbstract.Attributes.(*BranchStore)
                println("BranchId: ", v.BranchID.ToString())
                println("The counter is: ", innerStore.Counter)
                println("The counter at rewinded time is: ", rewindStore.Counter)
            }
        case strings.HasPrefix(line, "rewind"):
            line := strings.TrimSpace(line[6:])
            number, err := strconv.Atoi(line)
            if nil != err {
                log.Println("Can only rewind to a number")
                break
            }
            println("Rewinding to: ", uint64(number))
            Dispatch(dispatcher, event.Windback(uint64(number), uint64(time.Now().Unix()), store.NextID()))
        case line == "branch":
            branchID, err := AskForBranchID(l, timeStore)
            if nil != err {
                panic(err)
            }
            branch := timeStore.Branches[branchID]
            store := timeStore.Stores[branch.StoreID]
            // var lastEventID event.ID = event.ZeroID
            lastEventID := store.LastEvent.ID()
            Dispatch(dispatcher, event.NewBranch(branchID, lastEventID, uint64(time.Now().Unix()), store.NextID()))
        }
    }
}

