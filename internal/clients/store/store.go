// Package store contains the implementation for the store used by various internal services to
// preserve state, both static and dynamic as needed.
//
// It will permit updates, retrievals and will generate events on updates to registered subscribers.
//
// The etcd clientv3 package documentation can be found at https://pkg.go.dev/go.etcd.io/etcd/clientv3?tab=doc
//
// The primary methods to interact with the backing store (etcd at present) provided here are
//
//		Write()
//		WriteMultiple()
//		Read()
//		ReadMultiple()
//		ReadWithPrefix()
//		Delete()
//		DeleteMultiple()
//		DeleteWithPrexix()
//
// which typically take a string key or set of keys along with a set of string values for the
// WriteXxx() methods.
//
// The ReadXxx() methods areturn an array of KeyValueResponse structs which describe a set
// of zero of more key/value pairs. Note that the keys are returned as string but the values
// are []byte slices. These can readily be converted to strings as necessary.
//
// There are a set of methods used to establish a Store context object via the New() or
// NewWithDefault(). Alternatively the Store object can be directly allocated/initialized
// in a conventional fashion. The following additional methods mya also prove useful
//
//		Initialize()
//		SetAddress()
//		SetTimeoutConnect()
//		SetTimeoutRequest()
//
//
// Once a Store object has been created and initialized, a connection needs to be established
// with the backend databased before any IO can take place, and once the store is no longer
// required, the connection must be released. These operations can be achieved with the
// following methods
//
//		Connect()
//		Disconnect()
//
//
// Namespace
//
// The store package will prepend all keys with a constant prefix name to allow CloudChamber
// to share an etcd instance is that is require, though at least for the present this is
// not advised.
//
// Currently this prefix is the string
//
//		/CloudChamber/v0.1
//
// This prefix is not visible to any client of the store package but can be useful with any
// maintenance operations are being undertaken on the etcd store using (say) the etcdctl
// utility. In particular all CloudChamber can be listed using
//
//		%gopath%\bin\etcdctl get --write-out=simple --prefix /CloudChamber/v0.1
//
// and all the keys can be removed with
//
//		%gopath%\bin\etcdctl del --prefix /CloudChamber/v0.1
//
//
//
package store

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Jim3Things/CloudChamber/internal/config"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/namespace"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// All CloudChamber key-value pairs are stored under a common namespace root.
// This is fixed and cannot be changed programatically. However, to help
// certain situations, such as test, an additional suffix to the namespace
// can be added, changed, removed etc.
//
// At no time can the root namespace be removed.
//
const (
	cloudChamberNamespace = string("/CloudChamber/v0.1")
)

// TraceFlags is the type used when setting of fetching the relevant flags
// representing the level of verbosity being used when tracing.
type TraceFlags uint

const (
	traceFlagEnabled TraceFlags = 1 << iota
	traceFlagExpandResults
)

type global struct {
	Stores     map[string]Store
	StoreMutex sync.Mutex

	DefaultTimeoutConnect  time.Duration
	DefaultTimeoutRequest  time.Duration
	DefaultEndpoints       []string
	DefaultTraceFlags      TraceFlags
	DefaultNamespaceSuffix string
}

// Store is a struct used to collect all the interesting state values associated with
// communicating with an instance of the store
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
	NamespaceSuffix   string
	TraceFlags        TraceFlags
}

// KeyValueArg is a struct used to describe one or more key/value pairs supplied
// to a call such as WriteMultiple()
//
type KeyValueArg struct {
	key   string
	value string
}

// KeyValueResponse is a struct used to describe one or more key/value pairs
// returned from a call to the store such as a ReadMultiple() or ReadPrefix()
// call.
//
type KeyValueResponse struct {
	key   string
	value []byte
}

var (
	storeRoot global

	// ErrStoreUnableToCreateClient indicates that it is not currently possible
	// to create a client.
	//
	ErrStoreUnableToCreateClient = errors.New("CloudChamber: unable to create a new client")

	// ErrStoreNotConnected indicates the store instance does not have a
	// currently active client. The Conect() method can be used to establist a client.
	//
	ErrStoreNotConnected = errors.New("CloudChamber: client not currently connected")

	// ErrStoreConnected indicates the request failed as the store is currently
	// connected and the request is not possible in that condition.
	//
	ErrStoreConnected = errors.New("CloudChamber: client currently connected")
)

// ErrStoreBadResultSize indicates the size of the result set does not match
// expectations. There may be either too many, or too few. Typically a single
// result way anticipated and more that that was received.
//
type ErrStoreBadResultSize struct {
	expected int
	actual   int
}

func (esbrs ErrStoreBadResultSize) Error() string {
	return fmt.Sprintf("CloudChamber: unexpected size for result set - got %v expected %v", esbrs.actual, esbrs.expected)
}

// ErrStoreKeyNotFound indicates the request key was not found when the store
// lookup/fetch was attempted.
//
type ErrStoreKeyNotFound string

func (esknf ErrStoreKeyNotFound) Error() string {
	return fmt.Sprintf("CloudChamber: key %q not found", string(esknf))
}

// ErrStoreNotImplemented indicated the called method does not yet have an
//implementation
//
type ErrStoreNotImplemented string

func (esni ErrStoreNotImplemented) Error() string {
	return fmt.Sprintf("CloudChamber: method %v not currently implemented", string(esni))
}

func (store *Store) traceEnabled() bool { return store.TraceFlags != 0 }

func (store *Store) trace(v TraceFlags) (result bool) {

	if v&traceFlagEnabled == traceFlagEnabled {
		return store.traceEnabled()
	}

	if v&store.TraceFlags != 0 {
		return true
	}

	return false
}

func getDefaultEndpoints() []string {
	return storeRoot.DefaultEndpoints
}

func getDefaultTimeoutConnect() time.Duration {
	return storeRoot.DefaultTimeoutConnect
}

func getDefaultTimeoutRequest() time.Duration {
	return storeRoot.DefaultTimeoutRequest
}

func getDefaultTraceFlags() TraceFlags {
	return storeRoot.DefaultTraceFlags
}

func getDefaultNamespaceSuffix() string {
	return storeRoot.DefaultNamespaceSuffix
}

func setDefaultNamespaceSuffix(suffix string) {
	storeRoot.DefaultNamespaceSuffix = suffix
}

func (store *Store) connected(op string) error {

	if nil != store.Client {
		err := ErrStoreConnected
		log.Printf("Unable to %v - store is connected - error: %v", op, err)
		return err
	}

	return nil
}

func (store *Store) disconnected(op string) error {

	if nil == store.Client {
		err := ErrStoreNotConnected
		log.Printf("Failed to %v - no current connection - error: %v", op, err)
		return err
	}

	return nil
}

// Initialize is a method used to initialise the basic global state used to access
// the back-end db service.
//
func Initialize(cfg *config.GlobalConfig) {

	storeRoot.DefaultEndpoints = []string{fmt.Sprintf("%s:%v", cfg.Store.EtcdService.Hostname, cfg.Store.EtcdService.Port)}
	storeRoot.DefaultTimeoutConnect = time.Duration(cfg.Store.ConnectTimeout) * time.Second
	storeRoot.DefaultTimeoutRequest = time.Duration(cfg.Store.RequestTimeout) * time.Second
	storeRoot.DefaultTraceFlags = TraceFlags(cfg.Store.TraceLevel)
	storeRoot.DefaultNamespaceSuffix = ""
}

// NewStore is a method to allocate a new Store struct using the
// defaults which can later be overridden with
//
//    SetAddress()
//    SetTimeoutConnection()
//    SetTimeoutResponse()
//
// providing the store has not yet connected to the back-end db service.
//
func NewStore() *Store {
	return New(
		getDefaultEndpoints(),
		getDefaultTimeoutConnect(),
		getDefaultTimeoutRequest(),
		getDefaultTraceFlags(),
		getDefaultNamespaceSuffix(),
	)
}

// New is a method supplied values. These values can later be overridden with
//
//    SetAddress()
//    SetTimeoutConnection()
//    SetTimeoutResponse()
//
// providing the store has not yet connected to the back-end db service.
//
func New(endpoints []string, timeoutConnect time.Duration, timeoutRequest time.Duration, traceFlags TraceFlags, namespace string) *Store {

	store := Store{
		Endpoints:       endpoints,
		TimeoutConnect:  timeoutConnect,
		TimeoutRequest:  timeoutRequest,
		TraceFlags:      traceFlags,
		NamespaceSuffix: namespace,
	}

	return &store
}

// Initialize is a method used to initialize a specific instance of a Store struct.
//
func (store *Store) Initialize(endpoints []string, timeoutConnect time.Duration, timeoutRequest time.Duration, traceFlags TraceFlags, namespaceSuffix string) error {

	if err := store.connected("Initialize"); err != nil {
		return err
	}

	store.SetAddress(endpoints)
	store.SetTimeoutConnect(timeoutConnect)
	store.SetTimeoutRequest(timeoutRequest)
	store.SetTraceFlags(traceFlags)
	store.SetNamespaceSuffix(namespaceSuffix)

	store.Client = nil

	return nil
}

func (store *Store) logEtcdResponseError(err error) {
	if store.traceEnabled() {
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
				log.Printf("bad cluster endpoints, which are not etcd servers: %v\n", err)
			}
		}
	}
}

// SetTraceFlags is a method that can be used to change the verbosity of the tracing.
//
// The verbosity level can be set or updated regardless of the connection state of the
// store object.
//
func (store *Store) SetTraceFlags(traceLevel TraceFlags) {

	store.TraceFlags = traceLevel
	return
}

// GetTraceFlags is a method to retrieve the current trace flags value.
//
func (store *Store) GetTraceFlags() TraceFlags { return store.TraceFlags }

// SetAddress is a method that can be used to set or update the set of one or more endpoint
// that CloudChamber should use to connect to the backend etcd store.
//
// This method can only be used when the store object is not connected.
//
func (store *Store) SetAddress(endpoints []string) (err error) {

	if err := store.connected("Set(address)"); err != nil {
		return err
	}

	store.Endpoints = endpoints
	return nil
}

// GetAddress is a method to retrieve the current set of addresses used to connect to
// the underlying store.
//
func (store *Store) GetAddress() []string { return store.Endpoints }

// SetTimeoutConnect is a method that can be used to update the timeout the store object
// uses when establishing a connection to the backend etcd store.
//
// This method can only be used when the store object is not connected.
//
func (store *Store) SetTimeoutConnect(timeout time.Duration) (err error) {

	if err := store.connected("Set(timeoutConnect)"); err != nil {
		return err
	}

	store.TimeoutConnect = timeout
	return nil
}

// GetTimeoutConnect is a method which can be used to query the current timeout being
// used when the connection to the underlying store service is being established.
//
func (store *Store) GetTimeoutConnect() time.Duration { return store.TimeoutConnect }

// SetTimeoutRequest is a method that can be used to update the timeout the store object
// uses when establishing a connection to the backend etcd store.
//
// The request timeout can be set or updated regardless of the connection state of the
// store object, although it will only affect IO initiated after the method has returned.
//
func (store *Store) SetTimeoutRequest(timeout time.Duration) (err error) {

	store.TimeoutRequest = timeout
	return nil
}

// GetTimeoutRequest is a method which can be used to query the current timeout being
// used for individual requests to the underlying store service.
//
func (store *Store) GetTimeoutRequest() time.Duration { return store.TimeoutRequest }

// SetNamespaceSuffix is a method that can be used to update the namespace prefix being
// used for all read and write operations. For production use this will typically be the
// default namespace but for test usage this is set to a specific test namespace to both
// avoid interfering with production data and to allow simpler cleanup and removal of
// old temporary test data.
//
func (store *Store) SetNamespaceSuffix(suffix string) error {

	const slash = "/"

	if err := store.connected("Set(namespaceSuffix)"); err != nil {
		return err
	}

	// remove any leading, or trailing "/" characters regardless of how many there might be.
	//
	suffix = strings.Trim(suffix, slash)

	if suffix == "" {
		store.NamespaceSuffix = ""
	} else {
		store.NamespaceSuffix = slash + suffix
	}
	return nil
}

// GetNamespaceSuffix is a method which can be used to query the current namespace prefix
// being used for individual requests to the underlying store service.
//
func (store *Store) GetNamespaceSuffix() string { return store.NamespaceSuffix }

// Connect is a method that will establish a connection between the store object and the
// backend etcd database. A connection is required before any IO to the database can be
// attempted.
//
func (store *Store) Connect() (err error) {

	if err := store.connected("CONNECT"); err != nil {
		return err
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   store.GetAddress(),
		DialTimeout: store.GetTimeoutConnect(),
	})

	if err != nil {
		log.Printf("Failed to establish connection to store - error: %v", err)
		return err
	}

	// Hookup the namespace prefixing mechanism
	//
	store.Client = cli
	store.UnprefixedKV = cli.KV
	store.UnprefixedLease = cli.Lease
	store.UnprefixedWatcher = cli.Watcher

	name := cloudChamberNamespace + store.GetNamespaceSuffix()

	cli.KV = namespace.NewKV(cli.KV, name)
	cli.Watcher = namespace.NewWatcher(cli.Watcher, name)
	cli.Lease = namespace.NewLease(cli.Lease, name)

	return nil
}

// Disconnect is a method used to terminate the connection between the store object
// instance and the backed etcd service. Once the connection has been terninated,
// no further IO should be attempted.
//
func (store *Store) Disconnect() {

	if nil == store.Client {
		log.Printf("Store is already disconnected. No action taken")
		return
	}

	store.Client.Close()

	store.Client = nil
	store.UnprefixedKV = nil
	store.UnprefixedLease = nil
	store.UnprefixedWatcher = nil

	return
}

// Cluster is a structure which describes aspects of a cluster and the members of that cluster.
//
type Cluster struct {
	ID      uint64
	Members []ClusterMember
}

// ClusterMember is a structure which describes aspecs of a single member within a cluster
//
type ClusterMember struct {
	ID         uint64
	Name       string
	PeerURLs   []string
	ClientURLs []string
}

// GetClusterMembers is a method to fetch a description of the cluster to which the store
// object is currently connected.
//
func (store *Store) GetClusterMembers() (result *Cluster, err error) {

	if err := store.disconnected("GetClusterMembers"); err != nil {
		return result, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest)
	response, err := store.Client.MemberList(ctx)
	cancel()

	if err != nil {
		store.logEtcdResponseError(err)
	} else {
		result = &Cluster{
			ID:      response.Header.GetClusterId(),
			Members: make([]ClusterMember, len(response.Members))}

		for i, member := range response.Members {
			result.Members[i] = ClusterMember{
				Name:       member.GetName(),
				ID:         member.GetID(),
				PeerURLs:   member.GetPeerURLs(),
				ClientURLs: member.GetClientURLs()}
		}

		if store.trace(traceFlagExpandResults) {
			for i, node := range result.Members {
				log.Printf("node [%v] Id: %v Name: %v\n", i, node.ID, node.Name)
				for i, url := range node.ClientURLs {
					log.Printf("  client [%v] URL: %v\n", i, url)
				}
				for i, url := range node.PeerURLs {
					log.Printf("  peer [%v] URL: %v\n", i, url)
				}
			}
		}

		log.Printf("Processed %v items", len(result.Members))
	}

	return result, err
}

// UpdateClusterConnections is a method which updates the current set of connections
// to the connected underlying store according to the currently available connections
//
func (store *Store) UpdateClusterConnections() error {

	if err := store.disconnected("UpdateClusterConnections"); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest)
	err := store.Client.Sync(ctx)
	cancel()

	if err != nil {
		store.logEtcdResponseError(err)
	}

	return err
}

// Write is a method to write a new value into the store or update an existing
// value for the supplied key.
//
// It is expected that the store will already have been initialized and connected
// to the backed db server.
//
func (store *Store) Write(key string, value string) error {

	if err := store.disconnected("WRITE"); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest)
	_, err := store.Client.Put(ctx, key, value)
	cancel()

	if err != nil {
		store.logEtcdResponseError(err)
	} else {
		log.Printf("wrote/updated key: %v value: %v", key, value)
	}

	return err
}

// WriteMultiple is a method to write or update a set of values using a supplied
// set of keys in a pair-wise fashion.
//
// This is essentially a convenience method to allow multiuple values to be fetched
// in a single call rather than repeating individual calls to the Write() method.
//
func (store *Store) WriteMultiple(keyValueSet []KeyValueArg) (err error) {

	var processedCount int

	if err = store.disconnected("WRITE(Multiple)"); err != nil {
		return err
	}

	// The timeout multiplier (5) is arbitrary. May not even be necessary.
	//
	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest*5)

	for _, vp := range keyValueSet {
		_, err = store.Client.Put(ctx, vp.key, vp.value)
		if err != nil {
			break
		}
		processedCount++
	}

	cancel()

	if err != nil {
		store.logEtcdResponseError(err)
		log.Printf("Unable to write all the key/value pairs - requested: %v achieved: %v", len(keyValueSet), processedCount)
	}

	if store.trace(traceFlagExpandResults) {
		for i := 0; i < processedCount; i++ {
			log.Printf("wrote/updated [%v/%v] key: %v value: %v", i, processedCount, keyValueSet[i].key, keyValueSet[i].value)
		}
	}

	log.Printf("Processed %v items", processedCount)

	return err
}

// Read is a method to read a single value from the store using the supplied key.
//
// It is expected that the store will already have been initialized and connected
// to the backed db server.
//
func (store *Store) Read(key string) (result []byte, err error) {

	if err = store.disconnected("READ"); err != nil {
		return result, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest)
	response, err := store.Client.Get(ctx, key)
	cancel()

	if err != nil {
		store.logEtcdResponseError(err)
	} else if 0 == len(response.Kvs) {
		err = ErrStoreKeyNotFound(key)
		log.Printf("unable to read the requested key/value pair - error: %v\n", err)
	} else if 1 != len(response.Kvs) {
		err = ErrStoreBadResultSize{1, len(response.Kvs)}
		log.Printf("expected a single result and instead received something else - error: %v\n", err)
	} else {
		result = response.Kvs[0].Value
		log.Printf("read key: %v value: %v", key, result)
	}

	return result, err
}

// ReadMultiple is a method to read a set of values from the store using a supplied
// set of keys is a pair-wise fashion.
//
// This is essentially a convenience method to allow multiuple values to be fetched
// in a single call rather than repeating individual calls to the Read() method.
//
func (store *Store) ReadMultiple(keySet []string) (results []KeyValueResponse, err error) {

	var processedCount int

	if err = store.disconnected("READ(Multiple)"); err != nil {
		return results, err
	}

	responses := make([]*clientv3.GetResponse, len(keySet))

	// The timeout multiplier (5) is arbitrary. May not even be necessary.
	//
	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest*5)

	for i, key := range keySet {
		responses[i], err = store.Client.Get(ctx, key)
		if err != nil {
			break
		}
		processedCount++
	}

	cancel()

	if err != nil {
		store.logEtcdResponseError(err)
		log.Printf("Unable to read all the key/value pairs - requested: %v achieved: %v", len(keySet), processedCount)
	} else {
		results = make([]KeyValueResponse, processedCount)

		for i := 0; i < processedCount; i++ {
			if 1 != len(responses[i].Kvs) {
				err = ErrStoreBadResultSize{processedCount, len(responses[i].Kvs)}
				log.Printf("number of responses did not match expectations - error: %v\n", err)
			} else {
				results[i].key = string(responses[i].Kvs[0].Key)
				results[i].value = responses[i].Kvs[0].Value
			}
		}
	}

	if store.trace(traceFlagExpandResults) {
		for i := 0; i < processedCount; i++ {
			log.Printf("read [%v/%v] key: %v value: %v", i, processedCount, string(results[i].key), string(results[i].value))
		}
	}

	log.Printf("Processed %v items", processedCount)

	return results, err
}

// ReadWithPrefix is a method used to query for a set of zero or more key/value pairs
// which have a common prefix. The method will return all matching key/value pairs so
// care should be taken with key naming to avoid attempting to fetch a large number
// of key/value pairs.
//
// It is not an error to attmept to retrieve an empty set. For example, when querying
// for the presence of a set of values, this method can be used which would successfully
// return an empty set of key/value pairs if there are no matches for the supplied key
// prefix.
//
func (store *Store) ReadWithPrefix(keyPrefix string) (result []KeyValueResponse, err error) {

	if err = store.disconnected("READ(Prefix)"); err != nil {
		return result, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest)
	response, err := store.Client.Get(ctx, keyPrefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	cancel()

	if err != nil {
		store.logEtcdResponseError(err)
	} else {
		result = make([]KeyValueResponse, len(response.Kvs))

		for i, kv := range response.Kvs {
			result[i] = KeyValueResponse{string(kv.Key), kv.Value}
		}

		if store.trace(traceFlagExpandResults) {
			for i, kv := range result {
				log.Printf("read [%v/%v] key: %v value: %v", i, len(result), string(kv.key), string(kv.value))
			}
		}

		log.Printf("Processed %v items", len(result))
	}

	return result, err
}

// Delete is a method used to remove a single key/value pair using the supplied name.
//
func (store *Store) Delete(key string) (err error) {

	if err = store.disconnected("DELETE"); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest)
	response, err := store.Client.Delete(ctx, key)
	cancel()

	if err != nil {
		store.logEtcdResponseError(err)
	} else if 0 == response.Deleted {
		err = ErrStoreKeyNotFound(key)
		log.Printf("failed to delete the requested key/value pair - error: %v\n", err)
	} else if 1 != response.Deleted {
		err = ErrStoreBadResultSize{1, int(response.Deleted)}
		log.Printf("expected a single deletion and instead received something else - error: %v\n", err)
	} else {
		log.Printf("deleted key: %v", key)
	}

	return err
}

// DeleteMultiple is a method that can be used to remove a set of key/value pairs.
//
// This is essentially a convenience method to allow multiuple values to be fetched
// in a single call rather than repeating individual calls to the Delete() method.
//
func (store *Store) DeleteMultiple(keySet []string) error {

	var err error
	var processedCount int

	if err = store.disconnected("DELETE(Multiple)"); err != nil {
		return err
	}

	// The timeout multiplier (5) is arbitrary. May not even be necessary.
	//
	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest*5)

	for _, key := range keySet {
		_, err = store.Client.Delete(ctx, key)
		if err != nil {
			break
		}
		processedCount++
	}

	cancel()

	if err != nil {
		store.logEtcdResponseError(err)
		log.Printf("Unable to delete all the keys - requested: %v achieved: %v", len(keySet), processedCount)
	}

	if store.trace(traceFlagExpandResults) {
		for i := 0; i < processedCount; i++ {
			log.Printf("deleted [%v/%v] key: %v", i, processedCount, keySet[i])
		}
	}

	log.Printf("Processed %v items", processedCount)

	return err
}

// DeleteWithPrefix is a method used to remove an entire sub-tree of key/value
// pairs which have a common key name prefix.
//
func (store *Store) DeleteWithPrefix(keyPrefix string) (err error) {

	if err = store.disconnected("DELETE(Prefix)"); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest)
	response, err := store.Client.Delete(ctx, keyPrefix, clientv3.WithPrefix())
	cancel()

	if err != nil {
		store.logEtcdResponseError(err)
	} else {
		log.Printf("deleted %v keys under prefix %v", response.Deleted, keyPrefix)
	}

	return err
}

// SetWatch is a method used to establish a watchpoint on a single key/value pari
//
func (store *Store) SetWatch(key string) error {
	return ErrStoreNotImplemented("SetWatch")
}

// SetWatchMultiple is a method used to establish a set of watchpoints on a set of
// key/value pairs.
//
// This is essentially a convenience method to allow multiuple values to be fetched
// in a single call rather than repeating individual calls to the SetWatch() method.
//
func (store *Store) SetWatchMultiple(key []string) error {
	return ErrStoreNotImplemented("SetWatchMultiple")
}

// SetWatchWithPrefix is a method used to establish a watchpoint on a entire
// sub-tree of key/value pairs whic have a common key name prefix/
//
func (store *Store) SetWatchWithPrefix(keyPrefix string) error {
	return ErrStoreNotImplemented("SetWatchWithPrefix")
}
