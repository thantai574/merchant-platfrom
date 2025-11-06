package errors

import (
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewErrorMsg(internal string, errcode uint32) error {
	return status.Error(codes.Code(errcode), internal)
}

var (
	// ErrPendingOrder will throw if transaction pending
	ErrPendingOrder = errors.New(" Giao dịch đang chờ xử lí")
	ErrFailOrder    = errors.New(" Giao dịch thất bại. Vui lòng thử lại sau")
	ErrPaidOrder    = errors.New(" Giao dịch đã được thanh toán")
	ErrExpiredOrder = errors.New(" Giao dịch đã hết thời gian thanh toán")
	ErrCancelOrder  = errors.New(" Giao dịch đã bị hủy")
	ErrGeneral      = errors.New("Đã có lỗi xảy ra. Vui lòng thử lại sau")
)
