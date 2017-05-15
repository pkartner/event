package main

import(
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"log"
	"time"	

	"github.com/boltdb/bolt"
)

type event struct {
	Count int
}

func main() {
	db, err := bolt.Open("test.db", 0600, nil)
	if nil != err {
		 panic(err)
	}
	defer db.Close()

	start := time.Now()
	serial := 10
	paralel := 100000
	readCount := 100

	for i := 0; i < serial; i++ {

		buffers := []bytes.Buffer{}
		for j := 0; j < paralel; j++ {
			e := &event{
				Count: i+j,
			}
			var buffer  bytes.Buffer
			enc := gob.NewEncoder(&buffer)
			if err := enc.Encode(e); nil != err {
				return 
			}
			buffers = append(buffers, buffer)
		}

		db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("eventbucket"))
			if nil != err {
				return err
			}
			for _, buffer := range buffers {
				id, err := b.NextSequence()
				if nil != err {
					return err
				}
				b.Put(itob(int(id)), buffer.Bytes())
			}
			return nil
		})
	}
	elapsed := time.Since(start)

	log.Printf("Wrote %d serial, %d paralel, %d total", serial, paralel, serial*paralel)
	log.Printf("Write took %s", elapsed)

	start = time.Now()
	values := [][]byte{}
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("eventbucket"))
		c := b.Cursor()

		_, v := c.First()
		for i := 0; i < readCount; i++{
			values = append(values, v)
			_, v = c.Next()
		}
		return nil
	})
	elapsed = time.Since(start)
	for _, v := range values {
		log.Print(v)
	}

	log.Printf("Read took %s", elapsed)
}

func itob(v int) []byte {
    b := make([]byte, 8)
    binary.BigEndian.PutUint64(b, uint64(v))
    return b
}