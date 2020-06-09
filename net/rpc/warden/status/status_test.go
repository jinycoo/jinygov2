package status

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.jd100.com/medusa/errors"
	"go.jd100.com/medusa/errors/pb"
)

func TestCodeConvert(t *testing.T) {
	var table = map[codes.Code]errors.Code{
		codes.OK: errors.OK,
		// codes.Canceled
		codes.Unknown:          errors.ServerErr,
		codes.InvalidArgument:  errors.RequestErr,
		codes.DeadlineExceeded: errors.Deadline,
		codes.NotFound:         errors.NothingFound,
		// codes.AlreadyExists
		codes.PermissionDenied:  errors.AccessDenied,
		codes.ResourceExhausted: errors.LimitExceed,
		// codes.FailedPrecondition
		// codes.Aborted
		// codes.OutOfRange
		codes.Unimplemented: errors.MethodNotAllowed,
		codes.Unavailable:   errors.ServiceUnavailable,
		// codes.DataLoss
		codes.Unauthenticated: errors.Unauthorized,
	}
	for k, v := range table {
		assert.Equal(t, toECode(status.New(k, "-500")), v)
	}
	for k, v := range table {
		assert.Equal(t, togRPCCode(v), k, fmt.Sprintf("togRPC code error: %d -> %d", v, k))
	}
}

func TestNoDetailsConvert(t *testing.T) {
	gst := status.New(codes.Unknown, "-2233")
	assert.Equal(t, toECode(gst).Code(), -2233)

	gst = status.New(codes.Internal, "")
	assert.Equal(t, toECode(gst).Code(), -500)
}

func TestFromError(t *testing.T) {
	t.Run("input general error", func(t *testing.T) {
		err := errors.New("general error")
		gst := FromError(err)

		assert.Equal(t, codes.Unknown, gst.Code())
		assert.Contains(t, gst.Message(), "general")
	})
	t.Run("input wrap error", func(t *testing.T) {
		err := errors.Wrap(errors.RequestErr, "hh")
		gst := FromError(err)

		assert.Equal(t, "-400", gst.Message())
	})
	t.Run("input ecode.Code", func(t *testing.T) {
		err := errors.RequestErr
		gst := FromError(err)

		//assert.Equal(t, codes.InvalidArgument, gst.Code())
		// NOTE: set all grpc.status as Unkown when error is ecode.Codes for compatible
		assert.Equal(t, codes.Unknown, gst.Code())
		// NOTE: gst.Message == str(ecode.Code) for compatible php leagcy code
		assert.Equal(t, err.Message(), gst.Message())
	})
	t.Run("input pb.Error", func(t *testing.T) {
		m := &timestamp.Timestamp{Seconds: time.Now().Unix()}
		detail, _ := ptypes.MarshalAny(m)
		err := &pb.Error{ErrCode: 2233, ErrMessage: "message", ErrDetail: detail}
		gst := FromError(err)

		assert.Equal(t, codes.Unknown, gst.Code())
		assert.Len(t, gst.Details(), 1)
		assert.Equal(t, "2233", gst.Message())
	})
	t.Run("input ecode.Status", func(t *testing.T) {
		m := &timestamp.Timestamp{Seconds: time.Now().Unix()}
		err, _ := errors.Error(errors.Unauthorized, "unauthorized").WithDetails(m)
		gst := FromError(err)

		//assert.Equal(t, codes.Unauthenticated, gst.Code())
		// NOTE: set all grpc.status as Unkown when error is ecode.Codes for compatible
		assert.Equal(t, codes.Unknown, gst.Code())
		assert.Len(t, gst.Details(), 2)
		details := gst.Details()
		assert.IsType(t, &pb.Error{}, details[0])
		assert.IsType(t, err.Proto(), details[1])
	})
}

func TestToEcode(t *testing.T) {
	t.Run("input general grpc.Status", func(t *testing.T) {
		gst := status.New(codes.Unknown, "unknown")
		ec := ToEcode(gst)

		assert.Equal(t, int(errors.ServerErr), ec.Code())
		assert.Equal(t, "-500", ec.Message())
		assert.Len(t, ec.Details(), 0)
	})

	t.Run("input pb.Error", func(t *testing.T) {
		m := &timestamp.Timestamp{Seconds: time.Now().Unix()}
		detail, _ := ptypes.MarshalAny(m)
		gst := status.New(codes.InvalidArgument, "requesterr")
		gst, _ = gst.WithDetails(&pb.Error{ErrCode: 1122, ErrMessage: "message", ErrDetail: detail})
		ec := ToEcode(gst)

		assert.Equal(t, 1122, ec.Code())
		assert.Equal(t, "message", ec.Message())
		assert.Len(t, ec.Details(), 1)
		assert.IsType(t, m, ec.Details()[0])
	})

	t.Run("input pb.Error and ecode.Status", func(t *testing.T) {
		gst := status.New(codes.InvalidArgument, "requesterr")
		gst, _ = gst.WithDetails(
			&pb.Error{ErrCode: 1122, ErrMessage: "message"},
			errors.SErrorf(errors.AccessKeyErr, "AccessKeyErr").Proto(),
		)
		ec := ToEcode(gst)

		assert.Equal(t, int(errors.AccessKeyErr), ec.Code())
		assert.Equal(t, "AccessKeyErr", ec.Message())
	})

	t.Run("input encode.Status", func(t *testing.T) {
		m := &timestamp.Timestamp{Seconds: time.Now().Unix()}
		st, _ := errors.SErrorf(errors.AccessKeyErr, "AccessKeyErr").WithDetails(m)
		gst := status.New(codes.InvalidArgument, "requesterr")
		gst, _ = gst.WithDetails(st.Proto())
		ec := ToEcode(gst)

		assert.Equal(t, int(errors.AccessKeyErr), ec.Code())
		assert.Equal(t, "AccessKeyErr", ec.Message())
		assert.Len(t, ec.Details(), 1)
		assert.IsType(t, m, ec.Details()[0])
	})
}
