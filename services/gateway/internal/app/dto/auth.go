package dto

type RegisterInput struct {
	Login       string
	Password    string
	FirstName   string
	LastName    string
	Patronymic  string
	Email       string
	PhoneNumber string
}

type LoginInput struct {
	Login    string
	Password string
}

type UserView struct {
	ID          string
	Login       string
	FirstName   string
	LastName    string
	Patronymic  string
	Email       string
	PhoneNumber string
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type AuthResult struct {
	User   UserView
	Tokens TokenPair
}
