package application

import (
	"context"
	"orders-system/domain/aggregates"
	"orders-system/proto/service_user"
)

func (application *OrderApplication) GetProfile(ctx context.Context, id string) (*aggregates.UserDetail, error) {
	entity_user, err := application.UserRepository.FindUserDetailById(ctx, &service_user.FindUserDetailByIdRequest{
		Id: id,
	})

	if err != nil {
		return nil, err
	}

	i := aggregates.UserDetail{}.ConvertUserDetailDTO2TypeUser(*entity_user.UserDetail)

	return i, nil
}
