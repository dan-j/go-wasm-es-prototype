package core

import (
	"encoding/json"
	"fmt"
)

type Event struct {
	Type string `json:"type"`

	raw json.RawMessage
}

func (e *Event) UnmarshalJSON(bytes []byte) error {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(bytes, &m); err != nil {
		return err
	}

	if t, ok := m["type"]; ok {
		var ts string
		if err := json.Unmarshal(t, &ts); err != nil {
			return err
		}

		e.Type = ts
	} else {
		return fmt.Errorf("invalid event structure, must contain 'type' as a string")
	}

	if data, ok := m["data"]; ok {
		e.raw = data
	}

	return nil
}

func (e Event) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type": e.Type,
		"data": e.raw,
	})
}

func (e *Event) SetData(data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	e.raw = bytes

	return nil
}

func (e Event) Data() (interface{}, error) {
	switch e.Type {
	case "InitialStateEvent":
		var event InitialStateEvent
		if err := json.Unmarshal(e.raw, &event); err != nil {
			return nil, err
		}
		return event, nil
	case "NameUpdatedEvent":
		var event NameUpdatedEvent
		if err := json.Unmarshal(e.raw, &event); err != nil {
			return nil, err
		}
		return event, nil
	case "ItemAddedEvent":
		var event ItemAddedEvent
		if err := json.Unmarshal(e.raw, &event); err != nil {
			return nil, err
		}
		return event, nil
	case "ItemDeletedEvent":
		var event ItemDeletedEvent
		if err := json.Unmarshal(e.raw, &event); err != nil {
			return nil, err
		}
		return event, nil
	case "ItemStatusUpdatedEvent":
		var event ItemStatusUpdatedEvent
		if err := json.Unmarshal(e.raw, &event); err != nil {
			return nil, err
		}
		return event, nil
	}

	return nil, fmt.Errorf("Event.Data(): unknown Type: %s", e.Type)
}

// InitialStateEvent is the first event sent over a new WS connection, clients should never emit this.
type InitialStateEvent struct {
	Thing Thing `json:"thing"`
}

type NameUpdatedEvent struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ItemAddedEvent struct {
	ThingID string `json:"thingId"`
	ItemID  string `json:"itemId"`
	Status  string `json:"status"`
}

type ItemDeletedEvent struct {
	ThingID string `json:"thingID"`
	ItemID  string `json:"itemId"`
}

type ItemStatusUpdatedEvent struct {
	ThingID string `json:"thingId"`
	ItemID  string `json:"itemId"`
	Status  string `json:"status"`
}
