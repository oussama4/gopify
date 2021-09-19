package gopify

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidToken     = errors.New("session token is Invalid")
	ErrTokenExpired     = errors.New("session token has expired")
	ErrSignatureInvalid = errors.New("session token signature is invalid")
)

// Payload holds the session token claims
type Payload struct {
	Iss  string `json:"iss"`
	Dest string `json:"dest"`
	Aud  string `json:"aud"`
	Sub  string `json:"sub"`
	Exp  int    `json:"exp"`
	Nbf  int    `json:"nbf"`
	Iat  int    `json:"iat"`
	Jti  string `json:"jti"`
	Sid  string `json:"sid"`
}

// SessionTtoekn is a session token that is used to authenticate an embedded Shopify app
type SessionToken struct {
	parts  []string
	apiKey string
	secret string
}

func NewSessionToken(token string, apiKey string, secret string) *SessionToken {
	parts := strings.Split(token, ".")
	st := &SessionToken{
		parts:  parts,
		apiKey: apiKey,
		secret: secret,
	}
	return st
}

// Verify checks the validity of the session token payload
func (pd *Payload) Verify(apiKey string) error {
	now := time.Now().Unix()
	if pd.Exp <= int(now) || pd.Exp == 0 {
		return ErrTokenExpired
	}
	if pd.Nbf >= int(now) || pd.Nbf == 0 {
		return ErrInvalidToken
	}
	if pd.Iat >= int(now) || pd.Iat == 0 {
		return ErrInvalidToken
	}
	if pd.Aud != apiKey {
		return ErrInvalidToken
	}
	return pd.validateShop()
}

func (pd *Payload) validateShop() error {
	iss, err := url.Parse(pd.Iss)
	if err != nil {
		return err
	}
	dest, err := url.Parse(pd.Dest)
	if err != nil {
		return err
	}
	if iss.Hostname() != dest.Hostname() {
		return ErrInvalidToken
	}
	re, _ := regexp.Compile(`[a-zA-Z0-9][a-zA-Z0-9-]*\.myshopify\.com`)
	if !re.MatchString(dest.Hostname()) {
		return ErrInvalidToken
	}
	return nil
}

// Decode decodes the session token payload
func (st *SessionToken) Decode() (*Payload, error) {
	payload, err := base64.RawURLEncoding.DecodeString(st.parts[1])
	if err != nil {
		return nil, err
	}

	p := &Payload{}

	if err := json.NewDecoder(bytes.NewBuffer(payload)).Decode(&p); err != nil {
		return nil, err
	}

	return p, nil
}

// VerifySignature verifies the signature of the session token
func (st *SessionToken) VerifySignature() error {
	headerAndPayload := strings.Join(st.parts[:2], ".")
	signature, err := base64.RawURLEncoding.DecodeString(st.parts[2])
	if err != nil {
		return err
	}
	hasher := hmac.New(sha256.New, []byte(st.secret))
	hasher.Write([]byte(headerAndPayload))
	if !hmac.Equal(hasher.Sum(nil), signature) {
		return ErrSignatureInvalid
	}

	return nil
}
