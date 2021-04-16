package frontend

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/Jim3Things/CloudChamber/simulation/internal/clients/timestamp"
	"github.com/Jim3Things/CloudChamber/simulation/internal/common"
	"github.com/Jim3Things/CloudChamber/simulation/internal/tracing"
	pb "github.com/Jim3Things/CloudChamber/simulation/pkg/protos/services"
)

func watchAddRoutes(routeBase *mux.Router) {
	router := routeBase.PathPrefix("/watch").Subrouter()

	router.HandleFunc("", handleWatch).Methods("GET")
}

// handleWatch processes a request to watch for any change in CloudChamber state
// that meets the supplied conditions.
func handleWatch(w http.ResponseWriter, r *http.Request) {
	// Get the optional query parameters.  Note that none of these are required,
	// and any missing parameters are set to values that should trigger
	// immediately.

	vars := r.URL.Query()

	// Get the simulated time criteria
	tick := getOr(vars, "tick", "-1")
	epoch := getOr(vars, "epoch", "0")

	// TODO: Handle the etcd store updates and trace log updates
	// rev := getOr(vars, "dbrev", "0")
	// logId := getOr(vars, "logId", "0")

	res := &pb.WatchResponse{
		EventType: &pb.WatchResponse_Expired{Expired: true},
	}

	ctx, span := tracing.StartSpan(context.Background(),
		tracing.WithName("Start watch request"),
		tracing.AsInternal(),
		tracing.WithContextValue(timestamp.EnsureTickInContext))
	defer span.End()

	err := doSessionHeader(
		ctx, w, r,
		func(ctx context.Context, session *sessions.Session) error {

			if err := server.sessions.ensureEstablishedSession(session); err != nil {
				return err
			}

			afterTick, err := ensureNumber("tick", tick)
			if err != nil {
				return err
			}

			afterEpoch, err := ensureNumber("epoch", epoch)
			if err != nil {
				return err
			}

			tracing.UpdateSpanName(
				ctx,
				"Watching for changes after tick:%d, epoch:%d",
				afterTick,
				afterEpoch)

			// Now wait for something to happen, or for the max time to wait to
			// expire.
			ctx, cancel := context.WithTimeout(ctx, server.watchTimeout)
			defer cancel()

			select {
			case data := <-timestamp.After(ctx, afterTick+1, afterEpoch+1):
				if data.Err != nil {
					return data.Err
				}

				res.EventType = &pb.WatchResponse_StatusResponse{StatusResponse: data.Status}
				break

			case <-ctx.Done():
				break
			}

			return nil
		})

	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	// The simulated time may have changed, so update the value in the active
	// context now.
	status, err := timestamp.Status(ctx)
	if err != nil {
		postHTTPError(ctx, w, err)
		return
	}

	ctx = common.ContextWithTick(ctx, status.Now)

	w.Header().Set("Content-Type", "application/json")

	p := jsonpb.Marshaler{}
	err = p.Marshal(w, res)

	httpErrorIf(ctx, w, err)

}

// getOr either returns the value associated withe key in the map, or it returns
// the supplied default value.
func getOr(m url.Values, key string, def string) string {
	if v, ok := m[key]; ok {
		switch len(v) {
		case 0:
			return def

		case 1:
			return v[0]

		default:
			// Set up the return value so that it will not parse, and will
			// result in the bad value error.
			return fmt.Sprintf("(%s)", strings.Join(v, ", "))
		}
	}

	return def
}
