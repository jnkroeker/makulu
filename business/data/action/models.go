package action

// Action represents an action and its coordinates
type Action struct {
	ID   string  `json:"id,omitempty"`
	Name string  `json:"name" validate:"required"`
	Lat  float64 `json:"lat" validate:"required"`
	Lng  float64 `json:"lng" validate:"required"`
	User string  `json:"user" validate:"required"`
}

// Action represents an action and its coordinates
type NewAction struct {
	Name string  `json:"name" validate:"required"`
	Lat  float64 `json:"lat" validate:"required"`
	Lng  float64 `json:"lng" validate:"required"`
	User string  `json:"user" validate:"required"`
}

// ==============================================================

type id struct {
	Resp struct {
		Entities []struct {
			ID string `json:"id"`
		} `json:"entities"`
	} `json:"resp"`
}

func (id) document() string {
	return `{
		entities: action {
			id
		}	
	}`
}
