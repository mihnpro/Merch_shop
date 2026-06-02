package contract

type RegisterRequest struct {
	Login       string `json:"login"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Patronymic  string `json:"patronymic,omitempty"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

type UserResponse struct {
	ID          string `json:"id"`
	Login       string `json:"login"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Patronymic  string `json:"patronymic,omitempty"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number,omitempty"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginResponse struct {
	User   UserResponse `json:"user"`
	Tokens Tokens       `json:"tokens"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}
