package types

type Session struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Pfp      string `json:"pfp"`
}

type CreateUser struct {
	Email    string
	Username string
	Pfp      string
}

type TransitSess struct {
	ID    string
	Email string
}
