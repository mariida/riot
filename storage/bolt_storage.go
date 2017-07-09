// Copyright 2017 ego authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package storage

import (
	"time"

	"github.com/boltdb/bolt"
)

var gwkDocuments = []byte("gwkDocuments")

type boltStorage struct {
	db *bolt.DB
}

// openBoltStorage open Bolt storage
func openBoltStorage(dbPath string) (Storage, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 3600 * time.Second})
	// db, err := bolt.Open(dbPath, 0600, &bolt.Options{})
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(gwkDocuments)
		return err
	})
	if err != nil {
		db.Close()
		return nil, err
	}
	return &boltStorage{db}, nil
}

// WALName returns the path to currently open database file.
func (s *boltStorage) WALName() string {
	return s.db.Path()
}

// Set executes a function within the context of a read-write managed
// transaction. If no error is returned from the function then the transaction
// is committed. If an error is returned then the entire transaction is rolled back.
// Any error that is returned from the function or returned from the commit is returned
// from the Update() method.
func (s *boltStorage) Set(k []byte, v []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(gwkDocuments).Put(k, v)
	})
}

// Get executes a function within the context of a managed read-only transaction.
// Any error that is returned from the function is returned from the View() method.
func (s *boltStorage) Get(k []byte) (b []byte, err error) {
	err = s.db.View(func(tx *bolt.Tx) error {
		b = tx.Bucket(gwkDocuments).Get(k)
		return nil
	})
	return
}

// Delete deletes a key. Exposing this so that user does not
// have to specify the Entry directly.
func (s *boltStorage) Delete(k []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(gwkDocuments).Delete(k)
	})
}

// ForEach get all key and value
func (s *boltStorage) ForEach(fn func(k, v []byte) error) error {
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(gwkDocuments)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if err := fn(k, v); err != nil {
				return err
			}
		}
		return nil
	})
}

// Close releases all database resources. All transactions
// must be closed before closing the database.
func (s *boltStorage) Close() error {
	return s.db.Close()
}
