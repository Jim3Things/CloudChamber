package server

import (
    "context"
    "fmt"
    "strconv"

    "github.com/AsynkronIT/protoactor-go/actor"
    "github.com/golang/protobuf/ptypes/empty"
    "go.opentelemetry.io/otel/api/global"
    "go.opentelemetry.io/otel/api/kv"
    trc "go.opentelemetry.io/otel/api/trace"

    "github.com/Jim3Things/CloudChamber/internal/tracing"
    pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
    "github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

// +++ Logging interceptors

// Set up the logging context for an incoming message.  Establish the
// surrounding span, and associate it with this actor instance.  The
// span is terminated when the actor returns.
func ReceiveLogger(next actor.ReceiverFunc) actor.ReceiverFunc {
    return func(c actor.ReceiverContext, envelope *actor.MessageEnvelope) {
        tr := global.TraceProvider().Tracer("server")

        ctx, span := tr.Start(
            annotatedContext(context.Background(), envelope),
            fmt.Sprintf("Actor %q/Receive", c.Self()),
            trc.WithSpanKind(trc.SpanKindServer),
            trc.WithAttributes(kv.String(tracing.StackTraceKey, tracing.StackTrace())),
        )

        defer func() {
            ClearSpan(c.Self())
            span.End()
        }()

        SetSpan(c.Self(), span)

        hdr, msg, pid := actor.UnwrapEnvelope(envelope)

        Info(ctx, -1, fmt.Sprintf("Receive pid: %v, hdr: %v, msg: %v", pid, hdr, dumpMessage(msg)))

        next(c, envelope)
    }
}

// Log a send operation from an actor, using the span associated with that
// actor's instance
func SendLogger(next actor.SenderFunc) actor.SenderFunc {
    return func (c actor.SenderContext, target *actor.PID, envelope *actor.MessageEnvelope) {
        ctx := trc.ContextWithSpan(context.Background(), GetSpan(c.Self()))
        hdr, msg, pid := actor.UnwrapEnvelope(envelope)

        Info(ctx, -1, fmt.Sprintf("Sending pid: %v, hdr: %v, msg: %v", pid, hdr, dumpMessage(msg)))

        next(c, target, envelope)
    }
}

// Simple trace formatting for each of the known message types.
func dumpMessage(msg interface{}) string {
    switch msg := msg.(type) {
    case *actor.Stopping:
        return "(System) actor.Stopping"

    case *actor.Stopped:
        return "(System) actor.Stopped"

    case *actor.Restarting:
        return "(System) actor.Restarting"

    case *actor.PoisonPill:
        return "(System) actor.Started"


    case *actor.Started:
        return "(Infra) actor.Started"

    case *actor.Stop:
        return "(Infra) actor.Stop"

    case *actor.Watch:
        return "(Infra) actor.Watch"

    case *actor.Unwatch:
        return "(Infra) actor.Unwatch"

    case *actor.Terminated:
        return "(Infra) actor.Terminated"

    case *actor.Failure:
        return fmt.Sprintf("(Infra) actor.Failure: %v", msg)

    case *actor.Restart:
        return "(Infra) actor.Restart"


    case *common.Completion:
        return fmt.Sprintf("common.Completion error: %q", msg.Error)

    case *common.Timestamp:
        return fmt.Sprintf("common.Timestamp: Ticks: %d", msg.Ticks)

    case *pb.AutoStepRequest:
        return "AutoStepRequest"

    case *pb.PolicyRequest:
        return fmt.Sprintf(
            "PolicyRequest: policy: %v, delay: %ds.%dn, match: %d",
            msg.Policy, msg.MeasuredDelay.Seconds, msg.MeasuredDelay.Nanos, msg.MatchEpoch)

    case *pb.ResetRequest:
        return "ResetRequest"

    case *pb.NowRequest:
        return "NowRequest"

    case *pb.StepRequest:
        return "StepRequest"

    case *pb.DelayRequest:
        return fmt.Sprintf("DelayRequest: atLeast to %d, jitter %d", msg.AtLeast.Ticks, msg.Jitter)

    case *pb.GetStatusRequest:
        return "GetStatusRequest"

    case *pb.StatusResponse:
        return fmt.Sprintf(
            "StatusResponse: mode: %v, measured delay: %ds.%dn, current time: %d, number of waiters: %d, epoch: %d",
            msg.Policy, msg.MeasuredDelay.Seconds, msg.MeasuredDelay.Nanos, msg.Now.Ticks, msg.WaiterCount, msg.Epoch)

    case *empty.Empty:
        return "<empty>"

    default:
        m, ok := msg.(actor.SystemMessage)
        if ok {
            return fmt.Sprintf("(System) %v", m)
        }

        return fmt.Sprintf("%v", m)
    }
}

func annotatedContext(ctx context.Context, envelope *actor.MessageEnvelope) context.Context {
    parent := trc.SpanContext{
        TraceID:    trc.ID{},
        SpanID:     trc.SpanID{},
        TraceFlags: 0,
    }

    id, err := trc.SpanIDFromHex(envelope.Header.Get(tracing.SourceSpanID))
    if err == nil {
        parent.SpanID = id
    } else {
        return ctx
    }

    traceID, err := trc.IDFromHex(envelope.Header.Get(tracing.SourceTraceID))
    if err == nil {
        parent.TraceID = traceID
    } else {
        return ctx
    }

    flg, err := strconv.Atoi(envelope.Header.Get(tracing.SourceTraceFlgs))
    if err == nil {
        parent.TraceFlags = byte(flg)
    } else {
        return ctx
    }

    if parent.IsValid() {
        return trc.ContextWithRemoteSpanContext(ctx, parent)
    }

    return ctx
}
