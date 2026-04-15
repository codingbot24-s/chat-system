package rqrstype 

type SignUpBody struct {
	Name string `json:"name"`
	Email string `json:"email"`
	Password string `json:"password"`
}

type SignUpres struct {
	Success bool `json:"success"`
	Msg 	string `json:"message"`
	Token 	string  `json:"token"`
}



type LoginReq struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type LoginRes struct {
	Msg string `json:"message"`
	Token string `json:"token"`
}