package model

import "fmt"

type JsonResponser interface{}

type Result struct {
	Code int16
	Msg  string
}

type TokenResult struct {
	AccessToken  string
	RefreshToken string
}
type WhichResult struct {
	Which uint64
}

type TokenDetails struct {
	AccessToken  string
	RefreshToken string
	AccessUuid   string
	RefreshUuid  string
	AtExpires    int64
	RtExpires    int64
}

type AccessDetails struct {
	AccessToken string
	AccessUuid  string
}

type TokenCredential struct {
	Type  string
	Value string
}

type TokenCredentialEmptyErr struct{}

func (t *TokenCredentialEmptyErr) Error() string {
	return "No specified token credential."
}

type TokenInvalidAlgErr struct{ Alg interface{} }

func (t *TokenInvalidAlgErr) Error() string {
	return fmt.Sprintf("Unexpected signing method: %v", t.Alg)
}

type TokenExpiredErr struct{}

func (t *TokenExpiredErr) Error() string {
	return "Token has expired."
}

type TokenParseErr struct{ Msg interface{} }

func (t *TokenParseErr) Error() string {
	return fmt.Sprintf("Token parse error: %v", t.Msg)
}
