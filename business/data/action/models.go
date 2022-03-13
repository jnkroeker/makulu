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

type updateResult struct {
	UpdateAction struct {
		Action []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			Lat  string `json:"lat"`
			Lng  string `json:"lng"`
			User string `json:"user"`
		} `json:"action"`
		NumUids int `json:"numUids"`
	} `json:"updateAction"`
}

func (updateResult) document() string {
	return `{
		user {
			id
			name
			lat
			lng
			user
		}
	}`
}

type deleteResult struct {
	DeleteAction struct {
		Action []struct {
			ID    string `json:"id"`
			Email string `json:"email"`
			Name  string `json:"name"`
			Role  string `json:"role"`
		} `json:"action"`
		NumUids int    `json:"numUids"`
		Msg     string `json:"msg"`
	} `json:"deleteAction"`
}

func (deleteResult) document() string {
	return `{
		user {
			msg,
			numUids,
		}
	}`
}
