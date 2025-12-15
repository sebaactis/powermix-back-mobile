package mailer

type ContactRequest struct {
	Name     string `json:"name" validate:"required,min=6,max=30"`
	Email    string `json:"email" validate:"required,min=6,max=30,email"`
	Category string `json:"category" validate:"required"`
	Message  string `json:"message" validate:"required,min=20,max=250"`
}

type ContactResponse struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Category   string `json:"category"`
	Message    string `json:"message"`
	ApiMessage string `json:"apiMessage"`
}
