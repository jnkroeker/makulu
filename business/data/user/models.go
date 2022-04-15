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
// uses business/sys/validate package
type NewUser struct {
	Name            string `json:"name" validate:"required"`
	Email           string `json:"email" validate:"required"`
	Role            string `json:"role" validate:"required"`
	Password        string `json:"password" validate:"required"`
	PasswordConfirm string `json:"password_confirm" validate:"required"`
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
