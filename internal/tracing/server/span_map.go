package server

import (
    "sync"

    "github.com/AsynkronIT/protoactor-go/actor"
    "go.opentelemetry.io/otel/api/trace"
)

var spans = sync.Map{}
var parents = sync.Map{}

func GetSpan(pid *actor.PID) trace.Span {
    value, ok := spans.Load(pid)
    if !ok {
        return nil
    }
    return value.(trace.Span)
}

func ClearSpan(pid *actor.PID) {
    spans.Delete(pid)
}

func SetSpan(pid *actor.PID, span trace.Span) {
    spans.Store(pid, span)
}

func GetParentSpan(pid *actor.PID) trace.Span {
    value, ok := parents.Load(pid)
    if !ok {
        return nil
    }
    return value.(trace.Span)
}

func ClearParentSpan(pid *actor.PID) {
    parents.Delete(pid)
}

func SetParentSpan(pid *actor.PID, span trace.Span) {
    parents.Store(pid, span)
}

