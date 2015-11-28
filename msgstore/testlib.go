package msgstore

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/jeffjenkins/dispatchd/amqp"
	"reflect"
)

// ids to map
func idsToMap() {

}

// Check if the msg store is composed of exactly
func assertKeys(dbName string, keys map[int64]bool) error {
	// Open DB
	db, err := bolt.Open(dbName, 0600, nil)
	defer db.Close()
	if err != nil {
		return err
	}
	err = db.View(func(tx *bolt.Tx) error {
		// Check index
		bucket := tx.Bucket([]byte("message_index"))
		if bucket == nil {
			return nil
		}

		// get from db
		var indexKeys, err1 = keysForBucket(tx, "message_index")
		var contentKeys, err2 = keysForBucket(tx, "message_contents")
		if err1 != nil {
			return err1
		}
		if err2 != nil {
			return err2
		}

		// Check equality
		// TODO: return key diff
		indexNotKeys := subtract(indexKeys, keys)
		keysNotIndex := subtract(keys, indexKeys)

		if !reflect.DeepEqual(keys, indexKeys) {
			return fmt.Errorf("Different values in index!\nindexNotKeys:%q\nkeysNotIndex:%q", indexNotKeys, keysNotIndex)
		}
		if !reflect.DeepEqual(keys, contentKeys) {
			return fmt.Errorf("Different values in content!")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func subtract(original map[int64]bool, subtractThis map[int64]bool) []int64 {
	ret := make([]int64, 0)
	for id, _ := range original {
		_, found := subtractThis[id]
		if !found {
			ret = append(ret, id)
		}
	}
	return ret
}

func keysForBucket(tx *bolt.Tx, bucketName string) (map[int64]bool, error) {
	// Check index
	bucket := tx.Bucket([]byte(bucketName))
	if bucket == nil {
		return nil, fmt.Errorf("No bucket!")
	}
	var cursor = bucket.Cursor()
	var keys = make(map[int64]bool)
	for bid, _ := cursor.First(); bid != nil; bid, _ = cursor.Next() {
		fmt.Printf("%s, key:%d\n", bucketName, bytesToInt64(bid))
		keys[bytesToInt64(bid)] = true
	}
	return keys, nil
}

type TestResourceHolder struct {
}

func (trh *TestResourceHolder) AcquireResources(qm *amqp.QueueMessage) bool {
	return true
}
func (trh *TestResourceHolder) ReleaseResources(qm *amqp.QueueMessage) {

}
