package event

import(
    "encoding/binary"
    "bytes"
    "encoding/gob"

    "github.com/boltdb/bolt"
)

const EventBucket = "event"

type BoltEventStore struct {
    db *bolt.DB
    events chan *Event
} 

func NewBoltEventStore(db *bolt.DB) *BoltEventStore {
    boltEventStore := &BoltEventStore{
        events: make(chan *Event),
        db: db,
    }

    go func() {
        for {
            event := <- boltEventStore.events
            var buffer  bytes.Buffer
            enc := gob.NewEncoder(&buffer)
            if err := enc.Encode(event); nil != err {
                return 
            }
            db.Update(func(tx *bolt.Tx) error {
                b, err := tx.CreateBucketIfNotExists([]byte(EventBucket))
                if nil != err {
                    return err
                }
                id, err := b.NextSequence()
                if nil != err {
                    return err
                }
                b.Put(itob(int(id)), buffer.Bytes())
                return nil
            })
        }
    }()

    return boltEventStore
}

func (s *BoltEventStore) Add(e *Event) {
    s.events <- e
}

func itob(v int) []byte {
    b := make([]byte, 8)
    binary.BigEndian.PutUint64(b, uint64(v))
    return b
}

func (s *BoltEventStore) Restore(d *Dispatcher) error {
    err := s.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(EventBucket))
        c := b.Cursor()

        var event Event
    
        for k, v := c.First(); k != nil; k, v = c.Next() {
            var buffer bytes.Buffer
            dec := gob.NewDecoder(&buffer)
            buffer = *bytes.NewBuffer(v)
            err := dec.Decode(&event)
            if nil != err {
                return err
            }
            d.Handle(&event)
        }
        return nil
    })
    if nil != err {
        return err
    }
    return nil
}