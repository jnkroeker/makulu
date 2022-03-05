package user

// User represents someone with access to the system.
type User struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	PasswordHash string `json:"password_hash"`
}

// NewUser contains information needed to create a new User.
type NewUser struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Role            string `json:"role"`
	Password        string `json:"password"`
	PasswordConfirm string `json:"password_confirm"`
}

// =============================================================================

// everything in graphql has this json type. It requires this type of marshaling.
type addResult struct {
	AddUser struct {
		User []struct {
			ID string `json:"id"`
		} `json:"user"`
	} `json:"addUser"`
}

// these are the fields we want returned from the graphql query
func (addResult) document() string {
	return `{
		user {
			id
		}
	}`
}
