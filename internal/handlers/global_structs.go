package handlers

import "github.com/golang-jwt/jwt/v5"

type TokenClaims struct {
	Id string
	jwt.RegisteredClaims
}

type ChannelImage struct {
	Key     string
	Status  string
	ImageId uint
}

type ChannelDocument struct {
	Key        string
	Status     string
	DocumentId uint
}
