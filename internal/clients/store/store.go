// Package store contains the implementation for the store used by various internal services to
// preserve state, both static and dynamic as needed.
//
// It will permit updates, retrievals and will generate events on updates to registered subscribers.
//
// The etcd clientv3 package documentation can be found at https://pkg.go.dev/go.etcd.io/etcd/clientv3?tab=doc
//
package store

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	//    "github.com/etcd-io/etcd/clientv3"
	//    "github.com/etcd-io/etcd/etcdserver/api/v3rpc/rpctypes"

	//    "github.com/coreos/etcd/clientv3"
	//    "github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	//    "go.etcd.io/etcd/clientv3/namespace"
)

const (
	namespacePrefix = string("CloudChamber/v0.1/")
	defaultEndpoint = string("localhost:2379")

	defaultTimeoutConnect = 5 * time.Second
	defaultTimeoutRequest = 5 * time.Second
)

type global struct {
	Stores     map[string]Store
	StoreMutex sync.Mutex

	DefaultTimeoutConnect time.Duration
	DefaultTimeoutRequest time.Duration
	DefaultEndpoints      []string
}

// Store is a struct used to collect al the interesting state value associated with communicating with an instance of the store
//
type Store struct {
	Endpoints      []string
	TimeoutConnect time.Duration
	TimeoutRequest time.Duration
	Client         *clientv3.Client
	Mutex          sync.Mutex
}

var (
	storeRoot global

	defaultEndpoints = []string{defaultEndpoint}

	// ErrStoreInUse indicates the attempted operation failed as the underlying specified store was already in use.
	//
	ErrStoreInUse = errors.New("CloudChamber: store is currently in use")

	// ErrStoreUnableToCreateClient indicates that it is not currently possible to create a client.
	//
	ErrStoreUnableToCreateClient = errors.New("CloudChamber: unable to create a new client")

	// ErrStoreNoCurrentClient indicates the store instance does not have a currently active client. The Conect() method can be used to establist a client.
	//
	ErrStoreNoCurrentClient = errors.New("CloudChamber: no currently active client")
)

// Initialize is a method used to initialise the basic global state used to access the back-end db service.
//
func Initialize() {
	storeRoot.DefaultEndpoints = defaultEndpoints
	storeRoot.DefaultTimeoutConnect = defaultTimeoutConnect
	storeRoot.DefaultTimeoutRequest = defaultTimeoutRequest
}

// Initialize is a method used to initialize a specific instance of a Store struct.
//
func (store *Store) Initialize(endpoints []string, defaultTimeoutConnect time.Duration, defaultTimeoutRequest time.Duration) error {

	var err error

	/*
		store := Store(
			Name: name,
			DefaultTimeoutConnect: defaultTimeoutConnect,
			DefaultTimeoutRequest: defaultTimeoutRequest
		)

		if nil == store {
		    err := ErrEmptyKey
		    log.Printf("Unable to allocate new store instance - error: %v", err)
		    return nil,err
		}
	*/

	if nil != store.Client {

		err = ErrStoreInUse

		log.Printf("Unable to initialize a store that is already in use - error: %v", err)
		return err
	}

	store.Endpoints = endpoints
	store.TimeoutConnect = defaultTimeoutConnect
	store.TimeoutRequest = defaultTimeoutRequest

	store.Client = nil

	return nil
}

// NewWithDefaults is a method to allocate a new Store struct using the defaults which can later be overridden with
//
//    SetAddress()
//    SetTimeoutConnection()
//    SetTimeoutResponse()
//
// providing the store has not yet connect3d to the back-end db service.
//
func NewWithDefaults() (*Store, error) {
	return New(storeRoot.DefaultEndpoints, storeRoot.DefaultTimeoutConnect, storeRoot.DefaultTimeoutRequest)
}

// New is a method supplied values. These values can later be overridden with
//
//    SetAddress()
//    SetTimeoutConnection()
//    SetTimeoutResponse()
//
// providing the store has not yet connect3d to the back-end db service.
//
func New(endpoints []string, timeoutConnect time.Duration, timeoutRequest time.Duration) (*Store, error) {

	var err error

	store := Store{
		Endpoints:      endpoints,
		TimeoutConnect: timeoutConnect,
		TimeoutRequest: timeoutRequest,
	}

	/*
		if nil == store {
			err = ErrStoreUnableToCreateClient

			log.Printf("Unable to allocate a new client instance - error: %v endpoints: %v", err, endpoints)
			return nil, err
		}
	*/
	//	err = store.Initialize(endpoints, timeoutConnect, timeoutRequest)

	return &store, err
}

// SetAddress is a method
//
func (store *Store) SetAddress(endpoints []string) error {

	if nil != store.Client {
		store.Client.Close()
		store.Client = nil
	}

	store.Endpoints = endpoints
	return nil
}

// SetTimeoutConnection is a method
//
func (store *Store) SetTimeoutConnection(timeout time.Duration) error {

	if nil != store.Client {
		store.Client.Close()
		store.Client = nil
	}

	store.TimeoutConnect = timeout
	return nil
}

// SetTimeoutRequest is a method
//
func (store *Store) SetTimeoutRequest(timeout time.Duration) error {

	if nil != store.Client {
		store.Client.Close()
		store.Client = nil
	}

	store.TimeoutRequest = timeout
	return nil
}

// Connect is a method
//
func (store *Store) Connect() error {

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   store.Endpoints,
		DialTimeout: store.TimeoutConnect,
	})

	if err != nil {
		log.Printf("Failed to establish connection to store - error: %v", err)
		return err
	}

	store.Client = cli

	return nil
}

// Disconnect is a method
//
func (store *Store) Disconnect() error {

	store.Client.Close()

	store.Client = nil

	return nil
}

// Write is a method
//
func (store *Store) Write(key string, value string) error {

	if nil == store.Client {
		err := ErrStoreNoCurrentClient
		log.Printf("Failed to write - no current connection - error: %v", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest)

	_, err := store.Client.Put(ctx, key, value)
	cancel()

	if err != nil {
		switch err {
		case context.Canceled:
			log.Printf("Write ctx is canceled by another routine: %v\n", err)

		case context.DeadlineExceeded:
			log.Printf("Write deadline exceeded: %v\n", err)

		case rpctypes.ErrEmptyKey:
			log.Printf("client-side error: %v\n", err)

		default:
			fmt.Printf("bad cluster endpoints, which are not etcd servers: %v\n", err)
		}
		return err
	}

	return nil
}

// Read is a method
//
func (store *Store) Read(key string) (string, error) {

	var err error
	var value string

	if nil == store.Client {
		err = ErrStoreNoCurrentClient

		log.Printf("Failed to read - no current connection - error: %v", err)
		return value, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest)

	response, err := store.Client.Get(ctx, key)
	cancel()

	if err != nil {
		switch err {
		case context.Canceled:
			log.Printf("ctx is canceled by another routine: %v\n", err)

		case context.DeadlineExceeded:
			log.Printf("ctx is attached with a deadline is exceeded: %v\n", err)

		case rpctypes.ErrEmptyKey:
			log.Printf("client-side error: %v\n", err)

		default:
			fmt.Printf("bad cluster endpoints, which are not etcd servers: %v\n", err)
		}
		return value, err
	}

	fmt.Printf("GET returned\n")
	for _, ev := range response.Kvs {
		value = string(ev.Value)
		fmt.Printf("%s: %s\n", ev.Key, ev.Value)
	}

	return value, nil
}

// SetWatch is a method
//
func (store *Store) SetWatch(key string) error {
	return nil
}

/*
func (store Store) UpdateEndpoints(endpoints []string) error {

	if nil != store.Client  {
		store.Client.SetEndpoints()
		store.Client = nil
	}

	store.Endpoints = endpoints

	ctx, cancel := context.WithTimeout(context.Background(), timeoutRequest)

	err := store.Sync(ctx)

	if nil != err {
		store.Disconnect()
		log.Printf("Failed to update endpoint(s) - connection closed - error: %v", err)
	}

	return err
}
*/
/*
func storeInitialize() error {

	storeRoot.DefaultStoreAddress = defaultStoreAddress
	storeRoot.DefaultTimeoutConnect = defaultTimeoutConnect
	storeRoot.DefaultTimeoutRequest = defaultTimeoutRequest

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{storeAddress},
		DialTimeout: timeoutConnect,
	})
	if err != nil {
		log.Printf("Failed to establish connection to standard store - error : %v", err)
		return err
	}

	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeoutRequest)

	_, err = cli.Put(ctx, "sample_key", "sample_value")
	cancel()
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	if err != nil {
		switch err {
		case context.Canceled:
			log.Printf("ctx is canceled by another routine: %v\n", err)
		case context.DeadlineExceeded:
			log.Printf("ctx is attached with a deadline is exceeded: %v\n", err)
		case rpctypes.ErrEmptyKey:
			log.Printf("client-side error: %v\n", err)
		default:
			fmt.Printf("bad cluster endpoints, which are not etcd servers: %v\n", err)
		}
		return err
	}
	// Output: client-side error: etcdserver: key is not provided

	return nil
}
*/
