package persist

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/gogo/protobuf/proto"
)

//
//            Persist
//

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

//
//                    Load
//

func LoadOne(db *bolt.DB, bucket string, key string, obj proto.Unmarshaler) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			return fmt.Errorf("Bucket not found: '%s'", bucket)
		}
		return LoadOneBoltTx(bucket, key, obj)
	})
}

func LoadOneBoltTx(bucket *bolt.Bucket, key string, obj proto.Unmarshaler) error {
	objBytes := bucket.Get([]byte(key))
	if objBytes == nil {
		return fmt.Errorf("Key not found: '%s'", key)
	}
	err := obj.Unmarshal(objBytes)
	if err != nil {
		return fmt.Errorf("Could not unmarshal key %s", key)
	}
	return nil
}

func LoadMany(db *bolt.DB, bucket string, objs map[string]proto.Unmarshaler) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil { // pragma: nocover
			return fmt.Errorf("create bucket: '%s'", bucket)
		}
		return LoadManyBoltTx(bucket, objs)
	})
}

func LoadManyBoltTx(bucket *bolt.Bucket, objs map[string]proto.Unmarshaler) error {
	for key, obj := range objs {
		err := LoadOneBoltTx(bucket, key, obj)
		if err != nil {
			return err
		}
	}
	return nil
}
