package auth

type User interface {
	// Username returns the users username.
	Username() string
	// FullName returns the users first and last name.
	FullName() string
	// Email returns the users email address.
	Email() string
	// Accounts returns a list of the users accounts.
	Accounts() []string
	// Roles returns a list of the users roles.
	Roles() []string
}
