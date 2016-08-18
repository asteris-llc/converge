// Copyright © 2016 Asteris, LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"golang.org/x/net/context"
)

const (
	// JWTLifetime is the acceptable lifetime of an issued JWT token
	JWTLifetime = 30 * time.Second

	// JWTAlg is the signing algorithm used for signing and verification
	JWTAlg = "HS512"
)

// JWTAuth does authentication between client and server
type JWTAuth struct {
	token []byte
}

// NewJWTAuth initializes a new JWTAuth from the token
func NewJWTAuth(token string) *JWTAuth {
	return &JWTAuth{token: []byte(token)}
}

// GetRequestMetadata gets the current request metadata
func (j *JWTAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := j.New()
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"authorization": "BEARER " + token,
	}, nil
}

// RequireTransportSecurity indicates whether JWT requires transport security
// (it does not)
func (j *JWTAuth) RequireTransportSecurity() bool { return false }

// New creates a signed token
func (j *JWTAuth) New() (string, error) {
	token := jwt.NewWithClaims(
		jwt.GetSigningMethod(JWTAlg),
		jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(JWTLifetime).Unix(),
		},
	)

	return token.SignedString(j.token)
}

// Verify a generated token
func (j *JWTAuth) Verify(material string) error {
	token, err := jwt.ParseWithClaims(
		material,
		&jwt.StandardClaims{},
		func(token *jwt.Token) (interface{}, error) {
			if alg, ok := token.Header["alg"]; !ok || alg != JWTAlg {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return j.token, nil
		},
	)
	if err != nil {
		return err
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		return errors.New("internal error, standard claims not present")
	}

	// standard verification: issued/expires at was not issued before now. No,
	// this doesn't account for clock skew. We'll see if it's actually a
	// problem.
	if !claims.VerifyIssuedAt(time.Now().Unix(), true) {
		return errors.New("issued at was invalid")
	}

	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return errors.New("expires at was invalid")
	}

	exp := time.Duration(claims.ExpiresAt) * time.Second
	iat := time.Duration(claims.IssuedAt) * time.Second

	if (exp - iat) != JWTLifetime {
		return fmt.Errorf("lifetime too large. Expected %s, was %s", JWTLifetime, (exp - iat))
	}

	return nil
}

func (j *JWTAuth) Protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var token string

		// Get token out of querystring, header, or cookie
		if query := r.URL.Query().Get("jwt"); query != "" {
			token = query
		} else if bearer := r.Header.Get("Authorization"); strings.HasPrefix(bearer, "BEARER ") {
			token = strings.TrimLeft(bearer, "BEARER ")
		} else if cookie, err := r.Cookie("jwt"); err != nil && cookie.Value != "" {
			token = cookie.Value
		}

		// Token is required
		if token == "" {
			http.Error(w, "authorization is required", http.StatusUnauthorized)
			return
		}

		// Validate token
		if err := j.Verify(token); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// looks like we're good, call the next handler
		next.ServeHTTP(w, r)
	})
}
