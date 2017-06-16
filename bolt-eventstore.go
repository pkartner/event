package event

import(
    "encoding/binary"
    "fmt"
    "bytes"
    "encoding/gob"

    "github.com/boltdb/bolt"
)

// EventBucket TODO
const EventBucket = "event"

// BoltEventStore TODO
type BoltEventStore struct {
    db *bolt.DB
    events chan Event
} 

// NewBoltEventStore TODO
func NewBoltEventStore(db *bolt.DB) *BoltEventStore {
    boltEventStore := &BoltEventStore{
        events: make(chan Event),
        db: db,
    }

    go func() {
        for {
            event := <- boltEventStore.events
            buffer := new(bytes.Buffer)
            enc := gob.NewEncoder(buffer)
            if err := enc.Encode(&event); nil != err {
                panic(err)
                return
            }
            err := db.Update(func(tx *bolt.Tx) error {
                b, err := tx.CreateBucketIfNotExists([]byte(EventBucket))
                if nil != err {
                    return err
                }
                b.Put(event.ID().Byte(), buffer.Bytes())
                return nil
            })
            if nil != err {
                panic(err)
            }
        }
    }()

    return boltEventStore
}

// Add TODO
func (s *BoltEventStore) Add(e Event) {
    s.events <- e
}

func itob(v int) []byte {
    b := make([]byte, 8)
    binary.BigEndian.PutUint64(b, uint64(v))
    return b
}

// Restore TODO
func (s *BoltEventStore) Restore(time uint64, handleFunc ReadEventHandleFunc) error {

   err := s.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(EventBucket))
        if b == nil {
            return fmt.Errorf("bucket %q not found", EventBucket)
        }
        c := b.Cursor()

        var event Event
    
        for k, v := c.First(); k != nil; k, v = c.Next() {
            buffer := bytes.NewBuffer(v)
            dec := gob.NewDecoder(buffer)
            err := dec.Decode(&event)
            if nil != err {
                return err
            }
            if event == nil {
                panic(fmt.Errorf("event is nil"))
            }
            if event.Time() > time {
                continue;
            }
            if err := handleFunc(event); nil != err {
                return err
            } 
        }
        return nil
    })
    return err
}