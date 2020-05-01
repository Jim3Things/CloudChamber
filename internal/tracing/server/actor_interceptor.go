package server

import (
    "context"
    "fmt"

    "github.com/AsynkronIT/protoactor-go/actor"
    "go.opentelemetry.io/otel/api/global"
    trc "go.opentelemetry.io/otel/api/trace"

    pb "github.com/Jim3Things/CloudChamber/pkg/protos/Stepper"
    "github.com/Jim3Things/CloudChamber/pkg/protos/common"
)

// Logging interceptors
func ReceiveLogger(next actor.ReceiverFunc) actor.ReceiverFunc {
    return func(c actor.ReceiverContext, envelope *actor.MessageEnvelope) {
        tr := global.TraceProvider().Tracer("server")

        ctx, span := tr.Start(
            context.Background(),
            fmt.Sprintf("Actor %q/Receive", c.Self()),
            trc.WithSpanKind(trc.SpanKindServer),
        )
        defer span.End()

        hdr, msg, pid := actor.UnwrapEnvelope(envelope)

        span.AddEvent(ctx, fmt.Sprintf("Receive pid: %v, hdr: %v, msg: %v", pid, hdr, dumpMessage(msg)))

        next(c, envelope)
        return
    }
}

func SendLogger(next actor.SenderFunc) actor.SenderFunc {
    return func (c actor.SenderContext, target *actor.PID, envelope *actor.MessageEnvelope) {
        tr := global.TraceProvider().Tracer("server")

        ctx, span := tr.Start(
            context.Background(),
            fmt.Sprintf("Actor %q/Send", c.Self()),
            trc.WithSpanKind(trc.SpanKindServer),
        )
        defer span.End()

        hdr, msg, pid := actor.UnwrapEnvelope(envelope)

        span.AddEvent(ctx, fmt.Sprintf("Receive pid: %v, hdr: %v, msg: %v", pid, hdr, dumpMessage(msg)))

        next(c, target, envelope)
        return
    }
}

func dumpMessage(msg interface{}) string {
    switch msg.(type) {
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
        return fmt.Sprintf("(Infra) actor.Failure: %v", msg.(*actor.Failure))

    case *actor.Restart:
        return "(Infra) actor.Restart"


    case *common.Completion:
        return "common.Completion"

    case *common.Timestamp:
        return "common.Timestamp"

    case *pb.AutoStepRequest:
        return "AutoStepRequest"

    case *pb.PolicyRequest:
        return fmt.Sprintf("PolicyRequest: %v", msg.(*pb.PolicyRequest))

    case *pb.ResetRequest:
        return "ResetRequest"

    case *pb.NowRequest:
        return "NowRequest"

    case *pb.StepRequest:
        return "StepRequest"

    case *pb.DelayRequest:
        dr := msg.(*pb.DelayRequest)
        return fmt.Sprintf("DelayRequest: atLeast to %d, jitter %d", dr.AtLeast.Ticks, dr.Jitter)

    default:
        m, ok := msg.(actor.SystemMessage)
        if ok {
            return fmt.Sprintf("(System) %v", m)
        } else {
            return fmt.Sprintf("%v", m)
        }
    }
}
