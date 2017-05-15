package event

import (
    "bytes"
    "encoding/gob"

    "github.com/boltdb/bolt"
)

// SnapshotBucket TODO
const SnapshotBucket = "snapshot"

// BoltSnapshotStore TODO
type BoltSnapshotStore struct {
    db *bolt.DB
}

// NewBoltSnapshotStore TODO
func NewBoltSnapshotStore(db *bolt.DB) *BoltSnapshotStore {
    return &BoltSnapshotStore{
        db: db,
    }
}

// Write TODO
func (w *BoltSnapshotStore) Write(s *Store) error {
    var buffer bytes.Buffer
    enc := gob.NewEncoder(&buffer)
    if err := enc.Encode(s); nil != err {
        return err
    }
    err := w.db.Update(func(tx *bolt.Tx) error {
        b, err := tx.CreateBucketIfNotExists([]byte(SnapshotBucket))
        if nil != err {
            return err
        }
        b.Put([]byte("test"), buffer.Bytes())
        return nil
    })
    if nil != err {
        return err
    }

    return nil
}

// Restore TODO
func (w *BoltSnapshotStore) Restore() (*Store, error) {
    var store Store
    err := w.db.View(func(tx *bolt.Tx) error {
        b := tx.Bucket([]byte(SnapshotBucket))
        value := b.Get([]byte("test"))
        buffer := bytes.NewBuffer(value)
        dec := gob.NewDecoder(buffer)
        if err := dec.Decode(&store); nil != err {
            return err
        }
        return nil
    })
    if nil != err {
        return nil, err
    }
    return &store, nil
}