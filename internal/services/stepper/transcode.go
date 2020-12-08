package stepper

// This file contains the functions to translate messages between the
// protobuf format used by grpc and the internal go-struct formats used by the
// stepper state machine.

import (
	"context"
	"math/rand"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"

	"github.com/Jim3Things/CloudChamber/internal/services/stepper/messages"
	"github.com/Jim3Things/CloudChamber/internal/sm"
	"github.com/Jim3Things/CloudChamber/pkg/protos/common"
	pb "github.com/Jim3Things/CloudChamber/pkg/protos/services"
)

// This module contains the functions that translate to and from the external
// protobuf message and the internal go-struct message formats.

// toInternal converts a protobuf message into the equivalent internal go
// structs.
func toInternal(
	ctx context.Context,
	msg interface{},
	ch chan *sm.Response) (sm.Envelope, error) {

	switch m := msg.(type) {
	case *pb.NowRequest:
		return messages.NewNow(ctx, ch), nil

	case *pb.PolicyRequest:
		return convertToInternalPolicyRequest(ctx, m, ch)

	case *pb.DelayRequest:
		return messages.NewDelay(ctx, calculateDueTime(m), ch), nil

	case *pb.AutoStepRequest:
		return messages.NewAutoStep(ctx, m.Epoch, ch), nil

	case *pb.StepRequest:
		return messages.NewStep(ctx, ch), nil

	case *pb.ResetRequest:
		return messages.NewReset(ctx, ch), nil

	case *pb.GetStatusRequest:
		return messages.NewGetStatus(ctx, ch), nil
	}

	return nil, errInvalidMessage
}

// +++ Return type conversions

// toExternalTimeStamp translates the response into a Timestamp protobuf
// message, or an error if required.
func toExternalTimeStamp(rsp *sm.Response) (*common.Timestamp, error) {
	if rsp.Err != nil {
		return nil, rsp.Err
	}

	return &common.Timestamp{Ticks: rsp.At}, nil
}

// toExternalStatusResponse translates teh response into a StatusResponse
// protobuf message, or an error if required.
func toExternalStatusResponse(rsp *sm.Response) (*pb.StatusResponse, error) {
	if rsp.Err != nil {
		return nil, rsp.Err
	}

	body, ok := rsp.Msg.(*messages.StatusResponseBody)
	if !ok {
		return nil, errInvalidMessage
	}

	return &pb.StatusResponse{
		Policy:        convertToExternalPolicy(body.Policy),
		MeasuredDelay: ptypes.DurationProto(body.MeasuredDelay),
		Now:           &common.Timestamp{Ticks: rsp.At},
		Epoch:         body.Guard,
		WaiterCount: body.Waiters,
	}, nil
}

// toExternalEmptyResponse translates the response into a well known empty
// protobuf message, or an error if required.
func toExternalEmptyResponse(rsp *sm.Response) (*empty.Empty, error) {
	if rsp.Err != nil {
		return nil, rsp.Err
	}

	return &empty.Empty{}, nil
}

// --- Return type conversions

// +++ Helper functions

// convertToInternalPolicyRequest converts a protobuf PolicyRequest into one of
// the internal policy messages.  This function changes policy option values
// into different messages, as that is used by the state machine to simplify
// the change-state processing.
func convertToInternalPolicyRequest(
	ctx context.Context,
	m *pb.PolicyRequest,
	ch chan *sm.Response) (sm.Envelope, error) {

	interval, err := ptypes.Duration(m.MeasuredDelay)
	if err != nil {
		return nil, err
	}

	switch m.Policy {
	case pb.StepperPolicy_NoWait:
		if interval != 0 {
			return nil, &errDelayMustBeZero{actual: interval}
		}

		return messages.NewNoWaitPolicy(ctx, m.MatchEpoch, ch), nil

	case pb.StepperPolicy_Measured:
		if interval <= 0 {
			return nil, &errDelayMustBePositive{actual: interval}
		}

		return messages.NewMeasuredPolicy(ctx, m.MatchEpoch, interval, ch), nil

	case pb.StepperPolicy_Manual:
		if interval != 0 {
			return nil, &errDelayMustBeZero{actual: interval}
		}

		return messages.NewManualPolicy(ctx, m.MatchEpoch, ch), nil
	}

	return nil, errInvalidMessage
}

// calculateDueTime constructs the due time for a delay request, accounting for
// any jitter value.  The concept of 'Jitter' does not proceed past this point.
func calculateDueTime(m *pb.DelayRequest) int64 {
	dueTime := m.AtLeast.Ticks

	if m.Jitter > 0 {
		dueTime += rand.Int63n(m.Jitter)
	}

	return dueTime
}

// --- Helper functions

// +++ Policy enum conversions
//     Note that the internal enum is used to hide the state machine state ID
//     that maps to a given policy option.

// convertToInternalPolicy is used to take an externally defined policy option
// and map it to an internal policy value.  This is used to decode the starting
// policy stored in the configuration file.
func convertToInternalPolicy(policy pb.StepperPolicy) int {
	switch policy {
	case pb.StepperPolicy_Invalid:
		return messages.PolicyInvalid

	case pb.StepperPolicy_NoWait:
		return messages.PolicyNoWait

	case pb.StepperPolicy_Measured:
		return messages.PolicyMeasured

	case pb.StepperPolicy_Manual:
		return messages.PolicyManual
	}

	return messages.PolicyInvalid
}

// convertToExternalPolicy converts an internal policy value into the protobuf
// enum value that can be used when responding to GetStatus grpc call.
func convertToExternalPolicy(policy int) pb.StepperPolicy {
	switch policy {
	case messages.PolicyInvalid:
		return pb.StepperPolicy_Invalid

	case messages.PolicyNoWait:
		return pb.StepperPolicy_NoWait

	case messages.PolicyMeasured:
		return pb.StepperPolicy_Measured

	case messages.PolicyManual:
		return pb.StepperPolicy_Manual
	}

	return pb.StepperPolicy_Invalid
}

// --- Policy enum conversions
