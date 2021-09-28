package gopify

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidToken     = errors.New("session token is Invalid")
	ErrTokenExpired     = errors.New("session token has expired")
	ErrSignatureInvalid = errors.New("session token signature is invalid")
	ErrNoTokenFound     = errors.New("no token found")
)

var (
	PayloadCtxKey = &contextKey{"TokenPayload"}
)

// contextKey is a value for use with context.WithValue.
type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return k.name
}

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

// checks the validity of the session token payload
func (g *Gopify) verifyPayload(pd *Payload) error {
	now := time.Now().Unix()
	if pd.Exp < int(now) || pd.Exp == 0 {
		return ErrTokenExpired
	}
	if pd.Nbf > int(now) || pd.Nbf == 0 {
		return ErrInvalidToken
	}
	if pd.Iat > int(now) || pd.Iat == 0 {
		return ErrInvalidToken
	}
	if pd.Aud != g.ApiKey {
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

// decodes the given session token and extracts the token payload from it
func (g *Gopify) DecodeSessionToken(token string) (*Payload, error) {
	parts := strings.Split(token, ".")
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	p := &Payload{}

	if err := json.NewDecoder(bytes.NewBuffer(payload)).Decode(&p); err != nil {
		return nil, err
	}

	err = g.verifyPayload(p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// VerifySignature verifies the signature of the session token
func (g *Gopify) VerifyTokenSignature(token string) error {
	parts := strings.Split(token, ".")
	headerAndPayload := strings.Join(parts[:2], ".")
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return err
	}
	hasher := hmac.New(sha256.New, []byte(g.ApiSecret))
	hasher.Write([]byte(headerAndPayload))
	if !hmac.Equal(hasher.Sum(nil), signature) {
		return ErrSignatureInvalid
	}

	return nil
}

func tokenFromHeader(r *http.Request) string {
	bearer := r.Header.Get("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}

// VerifyToken verifies the integrity of a session token
func (g *Gopify) VerifyToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := tokenFromHeader(r)
		if token == "" {
			http.Error(w, ErrNoTokenFound.Error(), http.StatusBadRequest)
			return
		}

		// decode token
		payload, err := g.DecodeSessionToken(token)
		if err == ErrInvalidToken {
			http.Error(w, ErrInvalidToken.Error(), http.StatusBadRequest)
			return
		} else if err == ErrTokenExpired {
			http.Error(w, ErrTokenExpired.Error(), http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// verify signature
		err = g.VerifyTokenSignature(token)
		if err == ErrSignatureInvalid {
			http.Error(w, ErrSignatureInvalid.Error(), http.StatusUnauthorized)
			return
		} else if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), PayloadCtxKey, payload)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
