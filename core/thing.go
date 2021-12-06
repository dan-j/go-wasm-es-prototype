package core

import (
	"fmt"
)

type Thing struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Items []Item `json:"items,omitempty"`
}

func (t *Thing) ApplyEvent(event Event) error {
	concrete, err := event.Data()
	if err != nil {
		return err
	}
	switch e := concrete.(type) {
	case InitialStateEvent:
		*t = e.Thing
	case NameUpdatedEvent:
		t.Name = e.Name
	case ItemAddedEvent:
		t.Items = append(t.Items, Item{
			ID:     e.ItemID,
			Status: e.Status,
		})
	case ItemDeletedEvent:
		for i, item := range t.Items {
			if item.ID == e.ItemID {
				copy(t.Items[i:], t.Items[i+1:])
				t.Items = t.Items[:len(t.Items)-1]
				break
			}
		}
	case ItemStatusUpdatedEvent:
		for i, item := range t.Items {
			if item.ID == e.ItemID {
				t.Items[i].Status = e.Status
				break
			}
		}
	default:
		return fmt.Errorf("unknown event: %T", e)
	}

	return nil
}

type Item struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}
