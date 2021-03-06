// Package store contains the implementation for the store used by various internal services to
// preserve state, both static and dynamic as needed.
//
// It will permit updates, retrievals and will generate events on updates to registered subscribers.
//
// The etcd clientv3 package documentation can be found at https://pkg.go.dev/go.etcd.io/etcd/clientv3?tab=doc
//
// The primary methods to interact with the backing store (etcd at present) provided here are
//
//		ListWithPrefix()
//		ReadTxn()
//		WriteTxn()
//		DeleteTxn()
//		DeleteWithPrefix()
//
// which typically take a string key or set of keys along with a set of string values for the
// WriteXxx() methods.
//
// The ReadXxx() methods return an array of KeyValueResponse structs which describe a set
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
	"fmt"
	"strings"
	"sync"
	"time"

	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	ns "go.etcd.io/etcd/clientv3/namespace"
	"go.etcd.io/etcd/etcdserver/api/v3rpc/rpctypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/config"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	"github.com/Jim3Things/CloudChamber/simulation/pkg/errors"
)

// All CloudChamber key-value pairs are stored under a common namespace root.
// This is fixed and cannot be changed programmatically. However, to help
// certain situations, such as test, an additional suffix to the namespace
// can be added, changed, removed etc.
//
// At no time can the root namespace be removed.
//
const (
	slash                 = "/"
	cloudChamberNamespace = string("/CloudChamber/v0.1")
	testNamespaceSuffix   = string("Test")
)

// TraceFlags is the type used when setting of fetching the relevant flags
// representing the level of verbosity being used when tracing.
type TraceFlags uint

const (
	traceFlagEnabled TraceFlags = 1 << iota
	traceFlagExpandResults
	traceFlagTraceKey
	traceFlagTraceKeyAndValue
	traceFlagExpandResultsInTest
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
	Namespace         string
	NamespaceSuffix   string
	TraceFlags        TraceFlags
}

var (
	storeRoot global
)

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

func (store *Store) connected(ctx context.Context) error {

	if nil != store.Client {
		return errors.ErrStoreConnected("already connected")
	}

	return nil
}

func (store *Store) disconnected(ctx context.Context) error {

	if nil == store.Client {
		return errors.ErrStoreNotConnected("already disconnected")
	}

	return nil
}

// Initialize is a method used to initialise the basic global state used to access
// the back-end db service.
//
func Initialize(ctx context.Context, cfg *config.GlobalConfig) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.OutsideTime),
		tracing.AsInternal(),
		tracing.WithName("Initialize store"))
	defer span.End()

	storeRoot.DefaultEndpoints = []string{
		fmt.Sprintf("%s:%d", cfg.Store.EtcdService.Hostname, cfg.Store.EtcdService.Port),
	}

	storeRoot.DefaultTimeoutConnect = time.Duration(cfg.Store.ConnectTimeout) * time.Second
	storeRoot.DefaultTimeoutRequest = time.Duration(cfg.Store.RequestTimeout) * time.Second
	storeRoot.DefaultTraceFlags = TraceFlags(cfg.Store.TraceLevel)
	storeRoot.DefaultNamespaceSuffix = ""

	tracing.Info(
		ctx,
		"EP: %v TimeoutConnect: %v TimeoutRequest: %v DefTrcFlags: %v NsSuffix: %v",
		storeRoot.DefaultEndpoints,
		storeRoot.DefaultTimeoutConnect,
		storeRoot.DefaultTimeoutRequest,
		storeRoot.DefaultTraceFlags,
		storeRoot.DefaultNamespaceSuffix)

	// See if any part of the test namespace requires initialization and optionally cleaning.
	//
	// NOTE: This only affects the test namespace, not the production namespace. Ideally it
	//       would not be here but the store is used by other services as part of their test
	//       operation and so this feature needs to be in common code.
	//
	PrepareTestNamespace(ctx, cfg)
}

// PrepareTestNamespace will prepare the store to be used for test purposes. Optionally, this will
// clean out the store of data left from previous runs and set a test specific namespace for all
// subsequent store operations.
//
// NOTE: It is expected that this is call soon after the store is initialized and before any
//       data related operations have taken place.
//
func PrepareTestNamespace(ctx context.Context, cfg *config.GlobalConfig) {
	if !cfg.Store.Test.UseTestNamespace {
		return
	}

	// It is meaningless to have both a unique per-instance test namespace
	// and to clean the store before the tests are run
	//
	if cfg.Store.Test.UseUniqueInstance && cfg.Store.Test.PreCleanStore {
		tracing.Fatal(
			ctx,
			"invalid configuration: : %v",
			errors.ErrStoreInvalidConfiguration("both UseUniqueInstance and PreCleanStore are enabled"))
	}

	// For test purposes, need to set an alternate namespace rather than
	// rely on the standard. From the configuration, we can either use the
	// standard, fixed, well-known prefix, or we can use a per-instance
	// unique prefix derived from the current time
	//
	testNamespace := testNamespaceSuffix

	if cfg.Store.Test.UseUniqueInstance {
		testNamespace += fmt.Sprintf("/%s", time.Now().Format(time.RFC3339Nano))

		tracing.UpdateSpanName(ctx, "Initialize store (using unique test namespace %q)", testNamespace)
	} else {
		testNamespace += "/Standard"

		tracing.UpdateSpanName(ctx, "Initialize store (using test namespace %q)", testNamespace)
	}

	if cfg.Store.Test.PreCleanStore {
		if err := cleanNamespace(ctx, testNamespace); err != nil {
			tracing.Fatal(ctx, "failed to pre-clean the store as requested - namespace: %s err: %v", testNamespace, err)
		}
	}

	setDefaultNamespaceSuffix(testNamespace)
}

func cleanNamespace(ctx context.Context, testNamespace string) error {
	store := NewStore()

	tracing.Info(ctx, "Starting store pre-clean of namespace %q", testNamespace)

	if store == nil {
		tracing.Fatal(ctx, "unable to allocate store context for pre-cleanup")
	}

	if err := store.SetNamespaceSuffix(""); err != nil {
		return err
	}

	if err := store.Connect(); err != nil {
		return err
	}

	if _, err := store.DeleteWithPrefix(ctx, testNamespace); err != nil {
		return err
	}

	store.Disconnect()

	return nil
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
func New(
	endpoints []string,
	timeoutConnect time.Duration,
	timeoutRequest time.Duration,
	traceFlags TraceFlags,
	namespace string) *Store {
	store := &Store{
		Endpoints:       endpoints,
		TimeoutConnect:  timeoutConnect,
		TimeoutRequest:  timeoutRequest,
		TraceFlags:      traceFlags,
		NamespaceSuffix: namespace,
	}

	return store
}

// Initialize is a method used to initialize a specific instance of a Store struct.
//
func (store *Store) Initialize(
	endpoints []string,
	timeoutConnect time.Duration,
	timeoutRequest time.Duration,
	traceFlags TraceFlags,
	namespaceSuffix string) error {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	if err := store.connected(ctx); err != nil {
		return err
	}

	if err := store.SetAddress(endpoints); err != nil {
		return err
	}

	if err := store.SetTimeoutConnect(timeoutConnect); err != nil {
		return err
	}

	if err := store.SetTimeoutRequest(timeoutRequest); err != nil {
		return err
	}

	store.SetTraceFlags(traceFlags)

	if err := store.SetNamespaceSuffix(namespaceSuffix); err != nil {
		return err
	}

	store.Client = nil
	return nil
}

func (store *Store) logEtcdResponseError(ctx context.Context, err error) {
	if store.traceEnabled() {
		switch err {
		case context.Canceled:
			_ = tracing.Error(ctx, "ctx is canceled by another routine: %v", err)

		case context.DeadlineExceeded:
			_ = tracing.Error(ctx, "ctx is attached with a deadline is exceeded: %v", err)

		case rpctypes.ErrEmptyKey:
			_ = tracing.Error(ctx, "client-side error: %v", err)

		default:
			if ev, ok := status.FromError(err); ok {
				code := ev.Code()
				if code == codes.DeadlineExceeded {
					// server-side context might have timed-out first (due to clock skew)
					// while original client-side context is not timed-out yet
					//
					_ = tracing.Error(ctx, "server-side deadline is exceeded: %v", code)
				}
			} else {
				_ = tracing.Error(ctx, "bad cluster endpoints, which are not etcd servers: %v", err)
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
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	store.TraceFlags = traceLevel
	tracing.Info(ctx, "TraceFlags: %v", store.GetTraceFlags())
}

// GetTraceFlags is a method to retrieve the current trace flags value.
//
func (store *Store) GetTraceFlags() TraceFlags { return store.TraceFlags }

// SetAddress is a method that can be used to set or update the set of one or more endpoint
// that CloudChamber should use to connect to the backend etcd store.
//
// This method can only be used when the store object is not connected.
//
func (store *Store) SetAddress(endpoints []string) error {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	if err := store.connected(ctx); err != nil {
		return err
	}

	store.Endpoints = endpoints

	tracing.Info(ctx, "EP: %v", store.GetAddress())
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
func (store *Store) SetTimeoutConnect(timeout time.Duration) error {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	if err := store.connected(ctx); err != nil {
		return err
	}

	store.TimeoutConnect = timeout

	tracing.Info(ctx, "TimeoutConnect: %v", store.GetTimeoutConnect())
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
func (store *Store) SetTimeoutRequest(timeout time.Duration) error {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	store.TimeoutRequest = timeout

	tracing.Info(ctx, "TimeoutRequest: %v", store.GetTimeoutRequest())
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
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.OutsideTime))
	defer span.End()

	if err := store.connected(ctx); err != nil {
		return err
	}

	// remove any leading, or trailing "/" characters regardless of how many there might be.
	//
	store.NamespaceSuffix = strings.Trim(suffix, slash)

	tracing.Info(ctx, "NamespaceSuffix: %v", store.GetNamespaceSuffix())
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
func (store *Store) Connect() error {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	if err := store.connected(ctx); err != nil {
		return err
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   store.GetAddress(),
		DialTimeout: store.GetTimeoutConnect(),
	})

	if err != nil {
		return errors.ErrStoreConnectionFailed{Endpoints: store.GetAddress(), Reason: err}
	}

	// Hookup the namespace prefixing mechanism
	//
	store.Client = cli
	store.UnprefixedKV = cli.KV
	store.UnprefixedLease = cli.Lease
	store.UnprefixedWatcher = cli.Watcher

	namespace := cloudChamberNamespace + slash

	suffix := store.GetNamespaceSuffix()

	if suffix != "" {
		namespace += suffix + slash
	}

	store.Namespace = namespace

	cli.KV = ns.NewKV(cli.KV, store.Namespace)
	cli.Watcher = ns.NewWatcher(cli.Watcher, store.Namespace)
	cli.Lease = ns.NewLease(cli.Lease, store.Namespace)

	return nil
}

// Disconnect is a method used to terminate the connection between the store object
// instance and the backed etcd service. Once the connection has been terminated,
// no further IO should be attempted.
//
func (store *Store) Disconnect() {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	if nil == store.Client {
		tracing.Info(ctx, "Store is already disconnected. No action taken")
		return
	}

	_ = store.Client.Close()

	store.Client = nil
	store.UnprefixedKV = nil
	store.UnprefixedLease = nil
	store.UnprefixedWatcher = nil
}

// Cluster is a structure which describes aspects of a cluster and the members of that cluster.
//
type Cluster struct {
	ID      uint64
	Members []ClusterMember
}

// ClusterMember is a structure which describes aspects of a single member within a cluster
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
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	if err = store.disconnected(ctx); err != nil {
		return nil, err
	}

	// Originally had a new context here - not sure if we can use the supplied
	// ctx or whether we still need the new one.
	//
	opCtx, cancel := context.WithTimeout(ctx, store.TimeoutRequest)
	response, err := store.Client.MemberList(opCtx)
	cancel()

	if err != nil {
		store.logEtcdResponseError(ctx, err)
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
				tracing.Info(ctx, "node [%v] Id: %v Name: %v", i, node.ID, node.Name)

				for j, url := range node.ClientURLs {
					tracing.Info(ctx, "  client [%v] URL: %v", j, url)
				}

				for k, url := range node.PeerURLs {
					tracing.Info(ctx, "  peer [%v] URL: %v", k, url)
				}
			}
		}

		tracing.Info(ctx, "Processed %v items", len(result.Members))
	}

	return result, err
}

// UpdateClusterConnections is a method which updates the current set of connections
// to the connected underlying store according to the currently available connections
//
func (store *Store) UpdateClusterConnections() error {
	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	if err := store.disconnected(ctx); err != nil {
		return err
	}

	opCtx, cancel := context.WithTimeout(context.Background(), store.TimeoutRequest)
	err := store.Client.Sync(opCtx)
	cancel()

	if err != nil {
		store.logEtcdResponseError(ctx, err)
	}

	return err
}

// Watch is a struct returned by a SetWatchXxx() call to the caller to allow
// the caller to observe the notification and when no longer required, close
// the channel via the Close() method
//
type Watch struct {
	key    string
	cancel context.CancelFunc
	Events chan WatchEvent
}

// WatchEventType is used to indicate the type of event that happened
// on the key under consideration
//
type WatchEventType uint

const (
	// WatchEventTypeCreate indicates the event was a Create operation
	//
	WatchEventTypeCreate WatchEventType = iota

	// WatchEventTypeUpdate indicates the event was an Update or modify operation
	//
	WatchEventTypeUpdate

	// WatchEventTypeDelete indicates the event was a Delete operation
	//
	WatchEventTypeDelete
)

// WatchEvent is a struct used to describe a specific event to a particular key
//
type WatchEvent struct {
	Type     WatchEventType
	Revision int64
	Key      string
	NewRev   int64
	NewVal   string
	OldRev   int64
	OldVal   string
}

func buildNotification(rev int64, ev *clientv3.Event) WatchEvent {
	if ev.IsCreate() {
		return WatchEvent{
			Type:     WatchEventTypeCreate,
			Revision: rev,
			Key:      string(ev.Kv.Key),
			OldRev:   RevisionInvalid,
			NewRev:   ev.Kv.CreateRevision,
			OldVal:   "",
			NewVal:   string(ev.Kv.Value),
		}
	} else if ev.IsModify() {
		return WatchEvent{
			Type:     WatchEventTypeUpdate,
			Revision: rev,
			Key:      string(ev.Kv.Key),
			OldRev:   ev.PrevKv.ModRevision,
			NewRev:   ev.Kv.ModRevision,
			OldVal:   string(ev.PrevKv.Value),
			NewVal:   string(ev.Kv.Value),
		}
	} else {
		return WatchEvent{
			Type:     WatchEventTypeDelete,
			Revision: rev,
			Key:      string(ev.Kv.Key),
			OldRev:   ev.PrevKv.ModRevision,
			NewRev:   RevisionInvalid,
			OldVal:   string(ev.PrevKv.Value),
			NewVal:   "",
		}
	}
}

// SetWatch is a method used to establish a watchpoint on a single key/value pair
//
func (store *Store) SetWatch(ctx context.Context, key string) (response *Watch, err error) {
	if err = store.disconnected(ctx); err != nil {
		tracing.Debug(ctx, "Attempt to SetWatch for key %q when store disconnected - error: %v", key, err)
		return nil, err
	}

	notifications := make(chan WatchEvent)

	opCtx, cancel := context.WithCancel(ctx)

	watchEvents := store.Client.Watch(opCtx, key, clientv3.WithPrevKV())

	go func ()  {
		for {
			select {
			case <-opCtx.Done():
				close(notifications)
				return

			default:
				for events := range watchEvents {
					for _, ev := range events.Events {
						notifications <- buildNotification(events.Header.GetRevision(), ev)
					}
				}
			}
		}
	}()

	response = &Watch{key: key, cancel: cancel, Events: notifications}

	return response, nil
}

// SetWatchWithPrefix is a method used to establish a watchpoint on a entire
// sub-tree of key/value pairs which have a common key name prefix
//
func (store *Store) SetWatchWithPrefix(ctx context.Context, keyPrefix string) (response *Watch, err error) {
	if err = store.disconnected(ctx); err != nil {
		tracing.Debug(ctx, "Attempt to SetWatch for key %q when store disconnected - error: %v", keyPrefix, err)
		return nil, err
	}

	notifications := make(chan WatchEvent)

	opCtx, cancel := context.WithCancel(ctx)

	watchEvents := store.Client.Watch(opCtx, keyPrefix, clientv3.WithPrevKV(), clientv3.WithPrefix())

	go func ()  {
		for {
			select {
			case <-opCtx.Done():
				close(notifications)
				return

			default:
				for events := range watchEvents {
					for _, ev := range events.Events {
						notifications <- buildNotification(events.Header.GetRevision(), ev)
					}
				}
			}
		}
	}()

	response = &Watch{key: keyPrefix, cancel: cancel, Events: notifications}

	return response, nil
}

// RevisionInvalid is returned from certain operations if
// failure cases and is also used when defining
// unconditional write to the store.
//
const (
	RevisionInvalid int64 = 0
)

// Condition is a type used to define the test to be
// applied when making a conditional write to the store
//
type Condition string

// Set of specifiers for the type of condition in a
// conditional record update.
//
const (
	ConditionCreate                 = Condition("new")
	ConditionRequired               = Condition("req")
	ConditionUnconditional          = Condition("*")
	ConditionRevisionNotEqual       = Condition("!=")
	ConditionRevisionLess           = Condition("<")
	ConditionRevisionLessOrEqual    = Condition("<=")
	ConditionRevisionEqual          = Condition("==")
	ConditionRevisionEqualOrGreater = Condition(">=")
	ConditionRevisionGreater        = Condition(">")
)

func (c *Condition) String() string { return string(*c) }

// Record is a struct defining a single value and the associated revision describing
// the store revision when the value was last updated.
//
type Record struct {
	Revision int64
	Value    string
}

func generatePrefetchKeys(req *Request) (*[]string, error) {
	prefetchKeys := make([]string, 0, len(req.Records))

	// Build an array of keys to supply as the arg to prefetch
	// on the WithPrefetch() option below
	//
	for k, r := range req.Records {
		c := req.Conditions[k]
		switch c {
		case ConditionUnconditional:
			// No need to prefetch if not comparing anything
			break

		case ConditionRevisionNotEqual:
			fallthrough

		case ConditionRevisionLess:
			fallthrough

		case ConditionRevisionLessOrEqual:
			fallthrough

		case ConditionRevisionEqual:
			fallthrough

		case ConditionRevisionEqualOrGreater:
			fallthrough

		case ConditionRevisionGreater:
			if r.Revision == RevisionInvalid {
				return nil, errors.ErrStoreBadArgRevision{Key: k, Requested: RevisionInvalid, Actual: r.Revision}
			}

			fallthrough

		case ConditionRequired:
			fallthrough

		case ConditionCreate:
			prefetchKeys = append(prefetchKeys, k)

		default:
			return nil, errors.ErrStoreBadArgCondition{Key: k, Condition: c.String()}
		}
	}

	return &prefetchKeys, nil
}

func checkConditions(stm concurrency.STM, req *Request) error {
	for k, r := range req.Records {
		c := req.Conditions[k]

		// If the revision is zero, we take that to mean the key
		// does not actually exist.
		//
		// Note that a key might well exist even if the value is
		// an empty string. At least I believe it can. We choose
		// to use the revision instead as a more reliable
		// indicator of existence.
		//
		rev := stm.Rev(k)

		switch c {
		case ConditionCreate:
			if rev != 0 {
				return errors.ErrStoreAlreadyExists(k)
			}

		case ConditionRequired:
			if rev == 0 {
				return errors.ErrStoreKeyNotFound(k)
			}

		case ConditionRevisionLess:
			if rev >= r.Revision {
				return errors.ErrStoreConditionFail{Key: k, Requested: r.Revision, Condition: c.String(), Actual: rev}
			}

		case ConditionRevisionLessOrEqual:
			if rev > r.Revision {
				return errors.ErrStoreConditionFail{Key: k, Requested: r.Revision, Condition: c.String(), Actual: rev}
			}

		case ConditionRevisionEqual:
			if rev != r.Revision {
				return errors.ErrStoreConditionFail{Key: k, Requested: r.Revision, Condition: c.String(), Actual: rev}
			}

		case ConditionRevisionNotEqual:
			if rev == r.Revision {
				return errors.ErrStoreConditionFail{Key: k, Requested: r.Revision, Condition: c.String(), Actual: rev}
			}

		case ConditionRevisionEqualOrGreater:
			if rev < r.Revision {
				return errors.ErrStoreConditionFail{Key: k, Requested: r.Revision, Condition: c.String(), Actual: rev}
			}

		case ConditionRevisionGreater:
			if rev <= r.Revision {
				return errors.ErrStoreConditionFail{Key: k, Requested: r.Revision, Condition: c.String(), Actual: rev}
			}

		case ConditionUnconditional:
			// do nothing

		default:
			return errors.ErrStoreBadArgCondition{Key: k, Condition: c.String()}
		}
	}

	return nil
}

// ListWithPrefix is a method used to query for a set of zero or more key/value pairs
// which have a common prefix. The method will return all matching key/value pairs so
// care should be taken with key naming to avoid attempting to fetch a large number
// of key/value pairs.
//
// It is not an error to attempt to retrieve an empty set. For example, when querying
// for the presence of a set of values, this method can be used which would successfully
// return an empty set of key/value pairs if there are no matches for the supplied key
// prefix.
//
func (store *Store) ListWithPrefix(ctx context.Context, keyPrefix string) (response *Response, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	if err = store.disconnected(ctx); err != nil {
		return nil, err
	}

	// We may choose to make use of listing just the keys on the Get() rquest by
	// adding clientv3.WithKeysOnly() to the list of applicable options.
	//
	opCtx, cancel := context.WithTimeout(ctx, store.TimeoutRequest)

	getResponse, err := store.Client.Get(
		opCtx,
		keyPrefix,
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend))
	cancel()

	if err != nil {
		store.logEtcdResponseError(ctx, err)
		return nil, err
	}

	resp := &Response{
		Revision: RevisionInvalid,
		Records:  make(map[string]Record, len(getResponse.Kvs)),
	}

	for i, kv := range getResponse.Kvs {
		key := string(kv.Key)
		val := string(kv.Value)
		rev := kv.ModRevision

		resp.Records[key] = Record{Revision: rev, Value: val}

		if store.trace(traceFlagExpandResults) {
			if store.trace(traceFlagTraceKeyAndValue) {
				tracing.Info(ctx, "read [%v/%v] key: %v rev: %v value: %q", i, len(getResponse.Kvs), key, rev, val)
			} else if store.trace(traceFlagTraceKey) {
				tracing.Info(ctx, "read [%v/%v] key: %v", i, len(getResponse.Kvs), key)
			}
		}
	}

	resp.Revision = getResponse.Header.GetRevision()

	tracing.Info(ctx, "Processed %v items", len(resp.Records))

	return resp, nil
}

// DeleteWithPrefix is a method used to remove an entire sub-tree of key/value
// pairs which have a common key name prefix.
//
func (store *Store) DeleteWithPrefix(ctx context.Context, keyPrefix string) (response *Response, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	if err = store.disconnected(ctx); err != nil {
		return nil, err
	}

	opCtx, cancel := context.WithTimeout(ctx, store.TimeoutRequest)
	opResponse, err := store.Client.Delete(opCtx, keyPrefix, clientv3.WithPrefix())
	cancel()

	if err != nil {
		store.logEtcdResponseError(ctx, err)
		return nil, err
	}

	resp := &Response{
		Revision: RevisionInvalid,
		Records:  make(map[string]Record),
	}

	resp.Revision = opResponse.Header.GetRevision()

	tracing.Info(ctx, "deleted %v keys under prefix %v", opResponse.Deleted, keyPrefix)

	return resp, nil
}

//
// ToDo: need to add WithAction(actionRoutine) and WithRawValue() options.
//

// ReadTxn is a method to fetch a set of arbitrary keys within a
// single txn so they form a (self-)consistent set.
//
func (store *Store) ReadTxn(ctx context.Context, request *Request) (response *Response, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	if err = store.disconnected(ctx); err != nil {
		return nil, err
	}

	var prefetchKeys *[]string

	if prefetchKeys, err = generatePrefetchKeys(request); err != nil {
		return nil, err
	}

	resp := &Response{
		Revision: RevisionInvalid,
		Records:  make(map[string]Record, len(request.Records)),
	}

	txnAction := func(stm concurrency.STM) error {

		if err = checkConditions(stm, request); err != nil {
			return err
		}

		// It is only now that we know the conditions have been
		// met for all the keys that we take the time to process
		// all the updates.
		//
		// If the revision is zero, we take that to mean the key
		// does not actually exist.
		//
		// Note that a key might well exist even if the value is
		// an empty string. At least I believe it can. We choose
		// to use the revision instead as a more reliable
		// indicator of existence.
		//
		for k := range request.Records {

			rev := stm.Rev(k)

			if rev != 0 {
				resp.Records[k] = Record{Revision: rev, Value: stm.Get(k)}
			}
		}

		return nil
	}

	txnResponse, err := concurrency.NewSTM(
		store.Client,
		txnAction,
		concurrency.WithIsolation(concurrency.ReadCommitted),
		concurrency.WithPrefetch(*prefetchKeys...),
	)

	if err != nil {
		return nil, err
	}

	if !txnResponse.Succeeded {
		return nil, errors.ErrStoreKeyReadFailure(request.Reason)
	}

	// And finally, the revision for each of the individual records, and
	// for the store as a whole.
	//
	resp.Revision = txnResponse.Header.GetRevision()

	for _, r := range resp.Records {
		r.Revision = resp.Revision
	}

	return resp, nil
}

// WriteTxn is a method to write/update a set of arbitrary keys within a
// single txn so they form a (self-)consistent set.
//
func (store *Store) WriteTxn(ctx context.Context, request *Request) (response *Response, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	if err = store.disconnected(ctx); err != nil {
		return nil, err
	}

	var prefetchKeys *[]string

	if prefetchKeys, err = generatePrefetchKeys(request); err != nil {
		return nil, err
	}

	resp := &Response{
		Revision: RevisionInvalid,
		Records:  make(map[string]Record),
	}

	txnAction := func(stm concurrency.STM) error {

		if err = checkConditions(stm, request); err != nil {
			return err
		}

		// It is only now that we know the conditions have been
		// met for all the keys that we take the time to process
		// all the updates.
		//
		for k, r := range request.Records {
			stm.Put(k, r.Value)
		}

		return nil
	}

	txnResponse, err := concurrency.NewSTM(
		store.Client,
		txnAction,
		concurrency.WithIsolation(concurrency.Serializable),
		concurrency.WithPrefetch(*prefetchKeys...),
	)

	if err != nil {
		return nil, err
	}

	if !txnResponse.Succeeded {
		return nil, errors.ErrStoreKeyWriteFailure(request.Reason)
	}

	// And finally, the revision for the store as a whole.
	//
	// NOTE: there are no individual records for each of the keys.
	//       By definition, the revision for each key is the same
	//       as that for the transaction as a whole.
	//
	resp.Revision = txnResponse.Header.GetRevision()

	return resp, nil
}

// DeleteTxn is a method to delete a set of arbitrary keys within a
// single txn so they form a (self-)consistent operation.
//
func (store *Store) DeleteTxn(ctx context.Context, request *Request) (response *Response, err error) {
	ctx, span := tracing.StartSpan(ctx,
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	if err = store.disconnected(ctx); err != nil {
		return nil, err
	}

	var prefetchKeys *[]string

	if prefetchKeys, err = generatePrefetchKeys(request); err != nil {
		return nil, err
	}

	resp := &Response{
		Revision: RevisionInvalid,
		Records:  make(map[string]Record),
	}

	txnAction := func(stm concurrency.STM) error {

		if err = checkConditions(stm, request); err != nil {
			return err
		}

		// it is only now that we know the conditions have been
		// met for all the keys that we take the time to process
		// all the updates.
		//
		for k := range request.Records {
			stm.Del(k)
		}

		return nil
	}

	txnResponse, err := concurrency.NewSTM(
		store.Client,
		txnAction,
		concurrency.WithIsolation(concurrency.Serializable),
		concurrency.WithPrefetch(*prefetchKeys...),
	)

	if err != nil {
		return nil, err
	}

	if !txnResponse.Succeeded {
		return nil, errors.ErrStoreKeyDeleteFailure(request.Reason)
	}

	// And finally, the revision for the store as a whole.
	//
	// NOTE: there are no individual records for each of the keys.
	//       By definition, the revision for each key is the same
	//       as that for the transaction as a whole.
	//
	resp.Revision = txnResponse.Header.GetRevision()

	return resp, nil
}
