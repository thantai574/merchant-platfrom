package application

import (
	"context"
	"orders-system/domain/constants"
	"orders-system/domain/entities"
	"orders-system/errors"
	"orders-system/proto/order_system"
	"orders-system/utils/helpers"
	"orders-system/utils/telegram"
	"time"
)

func (us *OrderApplication) FindByID(ctx context.Context, id string) (*entities.OrderEntity, error) {
	return us.OrderRepository.FindByID(ctx, id)
}

func (us *OrderApplication) InitOrder(ctx context.Context, order_dto *entities.OrderEntity) (*entities.OrderEntity, error) {
	if order_dto.LuckyMoneyID != "" {
		check, _ := us.OrderRepository.CheckLuckyMoney(ctx, order_dto.ToUserID, order_dto.LuckyMoneyID)
		if check != nil {
			return nil, errors.NewErrorMsg("Order already exits", 33)
		}
	}

	order_dto.CreatedAt = helpers.GetCurrentTime()
	order_dto.UpdatedAt = order_dto.CreatedAt

	expireTimeDuration := time.Duration(10)

	if order_dto.ExpireTime > 0 {
		expireTimeDuration = time.Duration(order_dto.ExpireTime)
		order_dto.ExpiredAt = order_dto.CreatedAt.Add(expireTimeDuration * time.Millisecond)
	} else {
		order_dto.ExpiredAt = order_dto.CreatedAt.Add(expireTimeDuration * time.Minute)
	}

	order_dto.Status.ConvertStatusOrderEntity(order_system.OrderStatus_ORDER_PENDING)

	initOrder, err := us.OrderRepository.Create(ctx, order_dto)

	return initOrder, err
}

func (us *OrderApplication) ProcessingOrder(ctx context.Context, order_dto *entities.OrderEntity) (*entities.OrderEntity, error) {
	order_dto.Status.ConvertStatusOrderEntity(order_system.OrderStatus_ORDER_PROCESSING)
	order_dto.UpdatedAt = helpers.GetCurrentTime()

	processOrder, err := us.OrderRepository.ProcessingOrderByID(ctx, order_dto)

	return processOrder, err
}

func (us *OrderApplication) SuccessOrder(ctx context.Context, order_dto *entities.OrderEntity) (*entities.OrderEntity, error) {
	order_dto.Status.ConvertStatusOrderEntity(order_system.OrderStatus_ORDER_SUCCESS)

	order_dto.UpdatedAt = helpers.GetCurrentTime()
	order_dto.SucceedAt = helpers.GetCurrentTime()
	successOrder, err := us.OrderRepository.ReplaceByID(ctx, order_dto)
	if err == nil {
		us.IPool.Submit(func() {
			_ = us.sendSuccessOrderTelegram(ctx, order_dto)
		})
	}

	return successOrder, err
}

func (us *OrderApplication) sendSuccessOrderTelegram(ctx context.Context, orderDto *entities.OrderEntity) error {
	if helpers.IsStringSliceContains([]string{constants.SUB_TRANSTYPE_WALLET_QR_STATIC,
		constants.SUB_TRANSTYPE_WALLET_WEB_TO_APP, constants.SUB_TRANSTYPE_WALLET_QR_DYNAMIC, constants.SUB_TRANSTYPE_WALLET_WEB_IN_APP},
		orderDto.SubOrderType) {
	} else {
		orderInfoSend := telegram.SendOrderInfo(*orderDto)

		var channelId int64

		switch orderDto.OrderType {
		case constants.TRANSTYPE_PAY_TO_MERCHANT, constants.TRANSTYPE_BEHALF_PAY_WALLET_TO_BANK, constants.TRANSTYPE_BEHALF_PAY_WALLET_TO_WALLET:
			channelId = us.Config.TelegramChannelId.QR
		case constants.TRANSTYPE_PAY_VA:
			channelId = us.Config.TelegramChannelId.Va

		default:
			channelId = us.Config.TelegramChannelId.QR
		}

		telegram.SendTelegram(orderInfoSend, channelId)
	}

	return nil
}

func (us *OrderApplication) CancelOrder(ctx context.Context, order_dto *entities.OrderEntity) (*entities.OrderEntity, error) {
	order_dto.Status.ConvertStatusOrderEntity(order_system.OrderStatus_ORDER_CANCEL)
	order_dto.UpdatedAt = helpers.GetCurrentTime()

	cancelOrder, err := us.OrderRepository.ReplaceByID(ctx, order_dto)

	return cancelOrder, err
}

func (us *OrderApplication) FailedOrder(ctx context.Context, order_dto *entities.OrderEntity) (*entities.OrderEntity, error) {
	order_dto.Status.ConvertStatusOrderEntity(order_system.OrderStatus_ORDER_FAILED)
	order_dto.UpdatedAt = helpers.GetCurrentTime()
	return us.OrderRepository.ReplaceByID(ctx, order_dto)
}

func (us *OrderApplication) VerifyingOrder(ctx context.Context, order_dto *entities.OrderEntity) (*entities.OrderEntity, error) {
	order_dto.Status.ConvertStatusOrderEntity(order_system.OrderStatus_ORDER_VERIFYING)
	verifyOrder, err := us.OrderRepository.ReplaceByID(ctx, order_dto)

	return verifyOrder, err
}
