package utils

type JwtWrapper struct {
	SecretKey       string
	Issuer          string
	ExpirationHours int64
}
