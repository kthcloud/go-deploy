package event_service

import "go-deploy/models/sys/event"

func Create(id string, params *event.CreateParams) error {
	return event.New().Create(id, params)
}
