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

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/namespace"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	namespacePrefix = string("/CloudChamber/v0.1")
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
	Endpoints         []string
	TimeoutConnect    time.Duration
	TimeoutRequest    time.Duration
	Mutex             sync.Mutex
	Client            *clientv3.Client
	UnprefixedKV      clientv3.KV
	UnprefixedWatcher clientv3.Watcher
	UnprefixedLease   clientv3.Lease
}

// KeyValue is a struct used to describe one or more key/value pairs returned on a read from the store
//
type KeyValue struct {
	key   string
	value string
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

	// ErrStoreBadResultSize indicates the size of the result set does not match expectations. There may be either too many, or too few. Typically a single
	// result way anticipated and more that that was received.
	//
	ErrStoreBadResultSize = errors.New("CloudChamber: unexpected size for result set")

	// ErrStoreNotImplemented indicated the called method does not yet have an implementation
	//
	ErrStoreNotImplemented = errors.New("CloudChamber: method not currently implemented")
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

func logEtcdResponseError(err error) {
	switch err {
	case context.Canceled:
		log.Printf("ctx is canceled by another routine: %v\n", err)

	case context.DeadlineExceeded:
		log.Printf("ctx is attached with a deadline is exceeded: %v\n", err)

	case rpctypes.ErrEmptyKey:
		log.Printf("client-side error: %v\n", err)

	default:
		if ev, ok := status.FromError(err); ok {
			code := ev.Code()
			if code == codes.DeadlineExceeded {
				// server-side context might have timed-out first (due to clock skew)
				// while original client-side context is not timed-out yet
				//
				log.Printf("server-side deadline is exceeded: %v\n", code)

			}
		} else {
			fmt.Printf("bad cluster endpoints, which are not etcd servers: %v\n", err)
		}
	}
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
	store.UnprefixedKV = cli.KV
	store.UnprefixedLease = cli.Lease
	store.UnprefixedWatcher = cli.Watcher

	cli.KV = namespace.NewKV(cli.KV, namespacePrefix)
	cli.Watcher = namespace.NewWatcher(cli.Watcher, namespacePrefix)
	cli.Lease = namespace.NewLease(cli.Lease, namespacePrefix)

	return nil
}

// Disconnect is a method
//
func (store *Store) Disconnect() error {

	store.Client.Close()

	store.Client = nil
	store.UnprefixedKV = nil
	store.UnprefixedLease = nil
	store.UnprefixedWatcher = nil

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
		logEtcdResponseError(err)
		return err
	}

	return nil
}

// WriteMultiple is a method
//
func (store *Store) WriteMultiple(keyValueSet []KeyValue) error {

	var err error

	if nil == store.Client {
		err = ErrStoreNoCurrentClient
		log.Printf("Failed to write - no current connection - error: %v", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest*5)
	for _, vp := range keyValueSet {
		_, err = store.Client.Put(ctx, vp.key, vp.value)
	}
	cancel()

	if err != nil {
		logEtcdResponseError(err)
		return err
	}

	return nil
}

// Read is a method
//
func (store *Store) Read(key string) (string, error) {

	var value string

	if nil == store.Client {
		err := ErrStoreNoCurrentClient

		log.Printf("Failed to read - no current connection - error: %v", err)
		return value, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest)
	response, err := store.Client.Get(ctx, key)
	cancel()

	if err != nil {
		logEtcdResponseError(err)
		return value, err
	}

	if 1 != len(response.Kvs) {
		err = ErrStoreBadResultSize
		log.Printf("expected a single result and instead received something else - error: %v expected: 1 received: %v\n", err, len(response.Kvs))
		return value, err
	}

	value = string(response.Kvs[0].Value)

	return value, nil
}

// ReadMultiple is a method
//
func (store *Store) ReadMultiple(keySet []string) ([]KeyValue, error) {

	var err error

	if nil == store.Client {
		err = ErrStoreNoCurrentClient
		log.Printf("Failed to read - no current connection - error: %v", err)
		return nil, err
	}

	responses := make([]*clientv3.GetResponse, len(keySet))

	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest*10)

	for i, key := range keySet {
		responses[i], err = store.Client.Get(ctx, key)
		if err != nil {
			break
		}
	}
	cancel()

	if err != nil {
		logEtcdResponseError(err)
		return nil, err
	}

	results := make([]KeyValue, len(responses))

	for i, ev := range responses {
		results[i].key = string(ev.Kvs[0].Key)
		results[i].value = string(ev.Kvs[0].Value)
	}

	return results, err
}

// ReadWithPrefix is a method
//
func (store *Store) ReadWithPrefix(keyPrefix string) ([]KeyValue, error) {

	var err error

	if nil == store.Client {
		err = ErrStoreNoCurrentClient
		log.Printf("Failed to read - no current connection - error: %v", err)
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest)

	response, err := store.Client.Get(ctx, keyPrefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	cancel()

	if err != nil {
		logEtcdResponseError(err)
		return nil, err
	}

	result := make([]KeyValue, len(response.Kvs))

	for i, kv := range response.Kvs {
		result[i] = KeyValue{string(kv.Key), string(kv.Value)}
	}

	return result, nil
}

// SetWatch is a method
//
func (store *Store) SetWatch(key string) error {
	return ErrStoreNotImplemented
}

// SetWatchWithPrefix is a method
//
func (store *Store) SetWatchWithPrefix(keyPrefix string) error {
	return ErrStoreNotImplemented
}
