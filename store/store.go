package store

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/floeit/floe/path"
)

// Store links events to the config rules
type Store interface {
	Save(key string, data interface{}) error
	Load(key string, thing interface{}) error
	// Event(event.Event)
}

// MemStore is a simple in memory key value store
type MemStore struct {
	sync.RWMutex
	stuff map[string]interface{}
}

// NewMemStore returns an initialised MemStore
func NewMemStore() *MemStore {
	return &MemStore{
		stuff: map[string]interface{}{},
	}
}

// Save saves the data at the key
func (m *MemStore) Save(key string, data interface{}) error {
	m.Lock()
	defer m.Unlock()
	m.stuff[key] = data
	return nil
}

// Load loads data from the key
func (m *MemStore) Load(key string, thing interface{}) error {
	m.RLock()
	defer m.RUnlock()
	d, ok := m.stuff[key]
	if !ok {
		return nil
	}
	// set the val of the pointer with the stored val
	val := reflect.Indirect(reflect.ValueOf(thing))
	sval := reflect.ValueOf(d)
	if val.Type() != sval.Type() {
		return errors.New("can not set mismatched types")
	}
	val.Set(sval)

	return nil
}

// LocalStore is a local disk store
type LocalStore struct {
	sync.RWMutex
	root  string
	stuff map[string]interface{}
}

// NewLocalStore returns a local store based at the root directory
func NewLocalStore(root string) (*LocalStore, error) {
	r, err := path.Expand(root)
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(r, 0700)
	if err != nil {
		return nil, err
	}

	return &LocalStore{
		root:  r,
		stuff: map[string]interface{}{},
	}, nil
}

// Save saves the data at the key
func (m *LocalStore) Save(key string, data interface{}) error {
	m.Lock()
	defer m.Unlock()
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	keyPath := filepath.Join(m.root, key) + ".json"
	err = ioutil.WriteFile(keyPath, b, 0644)
	if err != nil {
		if _, ok := err.(*os.PathError); !ok {
			return err
		}
		err = os.MkdirAll(filepath.Dir(keyPath), 0700)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(keyPath, b, 0644)
	}
	return nil
}

// Load loads data from the key
func (m *LocalStore) Load(key string, thing interface{}) error {
	m.RLock()
	defer m.RUnlock()
	keyPath := filepath.Join(m.root, key) + ".json"
	b, err := ioutil.ReadFile(keyPath)
	if err != nil {
		if _, ok := err.(*os.PathError); ok { // file not found is ok
			return nil
		}
		return err
	}
	return json.Unmarshal(b, thing)
}
