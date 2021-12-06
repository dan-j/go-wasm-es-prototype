package core

import "encoding/json"


func UnmarshalEvent(data []byte) (interface{}, error) {
	var et Event
	if err := json.Unmarshal(data, &et); err != nil {
		return nil, err
	}

	var event interface{}
	switch et.Type {
	case "NameUpdatedEvent":
		event = NameUpdatedEvent{}
	case "ItemAddedEvent":
		event = ItemAddedEvent{}
	case "ItemDeletedEvent":
		event = ItemDeletedEvent{}
	case "ItemStatusUpdatedEvent":
		event = ItemStatusUpdatedEvent{}
	}

	if err := json.Unmarshal(data, &event); err != nil {
		return nil, err
	}

	return event, nil
}
