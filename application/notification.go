package application

import (
	"fmt"
	"html"
	"orders-system/domain/entities"
	"orders-system/utils/helpers"
)

func (us *OrderApplication) CreateMessage(title, user_id, content string, data map[string]interface{}) error {
	_, err := us.IMessage.CreateMessage(entities.Message{
		Id:                   helpers.GetUUId(),
		Title:                title,
		Topic:                "/topics/" + user_id,
		Content:              content,
		ContentData:          "",
		NeedTitleTranslate:   false,
		NeedContentTranslate: false,
		StatusPush:           1,
		IsHidden:             false,
		CreatedAt:            helpers.GetCurrentTime(),
		UpdatedAt:            helpers.GetCurrentTime(),
	})

	device, err := us.IDevice.FindByUserId(user_id)
	if err != nil {
		return err
	}

	err = us.IFirebase.SendByToken([]string{device.DeviceToken}, fmt.Sprint(title), html.UnescapeString(content), data)

	return err
}
