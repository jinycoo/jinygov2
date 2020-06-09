package status

import (
	"strconv"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.jd100.com/medusa/errors"
	"go.jd100.com/medusa/errors/pb"
)

// togRPCCode convert ecode.Codo to gRPC code
func togRPCCode(code errors.Codes) codes.Code {
	switch code.Code() {
	case errors.OK.Code():
		return codes.OK
	case errors.RequestErr.Code():
		return codes.InvalidArgument
	case errors.NothingFound.Code():
		return codes.NotFound
	case errors.Unauthorized.Code():
		return codes.Unauthenticated
	case errors.AccessDenied.Code():
		return codes.PermissionDenied
	case errors.LimitExceed.Code():
		return codes.ResourceExhausted
	case errors.MethodNotAllowed.Code():
		return codes.Unimplemented
	case errors.Deadline.Code():
		return codes.DeadlineExceeded
	case errors.ServiceUnavailable.Code():
		return codes.Unavailable
	}
	return codes.Unknown
}

func toECode(gst *status.Status) errors.Code {
	gcode := gst.Code()
	switch gcode {
	case codes.OK:
		return errors.OK
	case codes.InvalidArgument:
		return errors.RequestErr
	case codes.NotFound:
		return errors.NothingFound
	case codes.PermissionDenied:
		return errors.AccessDenied
	case codes.Unauthenticated:
		return errors.Unauthorized
	case codes.ResourceExhausted:
		return errors.LimitExceed
	case codes.Unimplemented:
		return errors.MethodNotAllowed
	case codes.DeadlineExceeded:
		return errors.Deadline
	case codes.Unavailable:
		return errors.ServiceUnavailable
	case codes.Unknown:
		return errors.String(gst.Message())
	}
	return errors.ServerErr
}

// FromError convert error for service reply and try to convert it to grpc.Status.
func FromError(err error) *status.Status {
	err = errors.Cause(err)
	if code, ok := err.(errors.Codes); ok {
		// TODO: deal with err
		if gst, err := gRPCStatusFromEcode(code); err == nil {
			return gst
		}
	}
	gst, _ := status.FromError(err)
	return gst
}

func gRPCStatusFromEcode(code errors.Codes) (*status.Status, error) {
	var st *errors.Status
	switch v := code.(type) {
	// compatible old pb.Error remove it after nobody use pb.Error.
	case *pb.Error:
		return status.New(codes.Unknown, v.Error()).WithDetails(v)
	case *errors.Status:
		st = v
	case errors.Code:
		st = errors.FromCode(v)
	default:
		st = errors.Error(errors.Code(code.Code()), code.Message())
		for _, detail := range code.Details() {
			if msg, ok := detail.(proto.Message); ok {
				st.WithDetails(msg)
			}
		}
	}
	// gst := status.New(togRPCCode(st), st.Message())
	// NOTE: compatible with PHP swoole gRPC put code in status message as string.
	// gst := status.New(togRPCCode(st), strconv.Itoa(st.Code()))
	gst := status.New(codes.Unknown, strconv.Itoa(st.Code()))
	pbe := &pb.Error{ErrCode: int32(st.Code()), ErrMessage: gst.Message()}
	// NOTE: server return ecode.Status will be covert to pb.Error details will be ignored
	// and put it at details[0] for compatible old client
	return gst.WithDetails(pbe, st.Proto())
}

// ToEcode convert grpc.status to ecode.Codes
func ToEcode(gst *status.Status) errors.Codes {
	details := gst.Details()
	// reverse range details, details may contain three case,
	// if details contain pb.Error and ecode.Status use eocde.Status first.
	//
	// Details layout:
	// pb.Error [0: pb.Error]
	// both pb.Error and ecode.Status [0: pb.Error, 1: ecode.Status]
	// ecode.Status [0: ecode.Status]
	for i := len(details) - 1; i >= 0; i-- {
		detail := details[i]
		// compatible with old pb.Error.
		if pe, ok := detail.(*pb.Error); ok {
			st := errors.Error(errors.Code(pe.ErrCode), pe.ErrMessage)
			if pe.ErrDetail != nil {
				dynMsg := new(ptypes.DynamicAny)
				// TODO deal with unmarshalAny error.
				if err := ptypes.UnmarshalAny(pe.ErrDetail, dynMsg); err == nil {
					st, _ = st.WithDetails(dynMsg.Message)
				}
			}
			return st
		}
		// convert detail to status only use first detail
		if pb, ok := detail.(proto.Message); ok {
			return errors.FromProto(pb)
		}
	}
	return toECode(gst)
}
