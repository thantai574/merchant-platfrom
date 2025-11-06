package errors

import (
	"encoding/json"
	err_micro "github.com/micro/go-micro/v2/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RecoveryError(err interface{}) error {
	if i, ok := err.(*status.Status); ok {
		return status.Errorf(i.Code(), i.Message())
	} else if i, ok := err.(*err_micro.Error); ok {
		return status.Errorf(codes.Code(i.Code), i.GetDetail())
	}

	return status.Errorf(400, "%v", err)
}

func GetGrpcErrMessage(err error) string {
	if e, ok := status.FromError(err); ok {
		return e.Message()
	}

	if i, ok := err.(*err_micro.Error); ok {
		type CodeDetail struct {
			Code   int    `json:"code"`
			Detail string `json:"detail"`
		}
		var codeDetail CodeDetail
		err = json.Unmarshal([]byte(i.GetDetail()), &codeDetail)
		if err != nil {
			return i.Detail
		}
		return codeDetail.Detail
	}

	return err.Error()
}
