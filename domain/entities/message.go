package entities

import (
	"orders-system/domain/value_objects"
	"time"
)

type Message struct {
	Id                   string            `json:"message_id" bson:"_id"`
	Title                string            `json:"title" bson:"title"`
	Topic                string            `json:"topic" bson:"topic"`
	Content              string            `json:"content" bson:"content"`
	Uri                  value_objects.URI `json:"uri" bson:"uri"`
	ContentData          string            `json:"content_data" bson:"content_data"`
	NeedTitleTranslate   bool              `json:"need_title_translate" bson:"need_title_translate"`
	NeedContentTranslate bool              `json:"need_content_translate" bson:"need_content_translate"`
	IsHidden             bool              `json:"is_hidden" bson:"is_hidden"`
	Type                 string            `json:"type" bson:"type"`
	StatusPush           int32             `json:"status_push" bson:"status_push"`
	TimePush             string            `json:"time_push" bson:"time_push"`
	CreatedAt            time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at" bson:"updated_at"`
}
