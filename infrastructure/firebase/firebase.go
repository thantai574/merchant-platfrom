package firebase

import (
	"fmt"
	"gopkg.in/maddevsio/fcm.v1"
)

// Send Message 1 Topic

type IFirebase interface {
	Send(topic, title, body string, data map[string]interface{}) error
	SendByToken(tokens []string, title, body string, data map[string]interface{}) error
}

type repoImpl struct {
	FirebaseKey string
}

func NewRepoImpl(firebaseKey string) *repoImpl {
	return &repoImpl{FirebaseKey: firebaseKey}
}

func (r *repoImpl) Send(topic, title, body string, data map[string]interface{}) error {

	c := fcm.NewFCM(r.FirebaseKey)

	if topic[0:1] != "/" {
		topic = "/topics/" + topic
	}

	response, err := c.Send(fcm.Message{
		Data:             data,
		To:               topic,
		ContentAvailable: true,
		Priority:         fcm.PriorityHigh,
		Notification: fcm.Notification{
			Title: title,
			Body:  body,
		},
	})

	if err != nil {
		return err
	}

	fmt.Println("Title  :", title)
	fmt.Println("Body  :", body)
	fmt.Println("Status Code   :", response.StatusCode)
	fmt.Println("Success       :", response.Success)
	fmt.Println("Fail          :", response.Fail)
	fmt.Println("Canonical_ids :", response.CanonicalIDs)
	fmt.Println("Topic MsgId   :", response.MsgID)

	return nil
}

func (r *repoImpl) SendByToken(tokens []string, title, body string, data map[string]interface{}) error {

	c := fcm.NewFCM(r.FirebaseKey)

	response, err := c.Send(fcm.Message{
		Data:             data,
		RegistrationIDs:  tokens,
		ContentAvailable: true,
		Priority:         fcm.PriorityHigh,
		Notification: fcm.Notification{
			Title: title,
			Body:  body,
		},
	})

	fmt.Println("Title  :", title)
	fmt.Println("Body  :", body)
	if err != nil {
		fmt.Println("Title  :", err)
		return err
	}

	fmt.Println("Status Code   :", response.StatusCode)
	fmt.Println("Success       :", response.Success)
	fmt.Println("Fail          :", response.Fail)
	fmt.Println("Canonical_ids :", response.CanonicalIDs)
	fmt.Println("Topic MsgId   :", response.MsgID)

	return nil
}
