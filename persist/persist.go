package persist

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
)

func PersistOne(db *bolt.DB, bucket string, key string, obj proto.Marshaler) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil { // pragma: nocover
			return fmt.Errorf("create bucket: %s", err)
		}
		return PersistOneBoltTx(bucket, key, obj)
	})
}

func PersistOneBoltTx(bucket *bolt.Bucket, key string, obj proto.Marshaler) error {
	exBytes, err := obj.Marshal()
	if err != nil { // pragma: nocover -- no idea how to produce this error
		return fmt.Errorf("Could not marshal object")
	}
	return bucket.Put([]byte(key), exBytes)
}

func PersistMany(db *bolt.DB, bucket string, objs map[string]proto.Marshaler) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil { // pragma: nocover
			return fmt.Errorf("create bucket: %s", err)
		}
		return PersistManyBoltTx(bucket, objs)
	})
}

func PersistManyBoltTx(bucket *bolt.Bucket, objs map[string]proto.Marshaler) error {
	for key, obj := range objs {
		err := PersistOneBoltTx(bucket, key, obj)
		if err != nil {
			return err
		}
	}
	return nil
}
