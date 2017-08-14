package models

import "src/github.com/dgrijalva/jwt-go"

type Claim struct {
	User `json:"user"`
	jwt.StandardClaims
}

