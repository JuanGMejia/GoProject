package models

type User struct{
	Name string `json:"name"`
	Password string `json:"password,omitempty"`
	Role string `json:"role"`
}

