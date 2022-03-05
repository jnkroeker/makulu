package action

// Action represents an action and its coordinates
type Action struct {
	ID   string  `json:"id,omitempty"`
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
	User User    `json:"user"`
}

// User captures the User ID in the relationship with an Action
type User struct {
	ID string `json:"id"`
}

// NewAction contains information needed to create a new Action.
type NewAction struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
	User User    `json:"user"`
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
