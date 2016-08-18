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

package rpc_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/asteris-llc/converge/rpc"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTAuth(t *testing.T) {
	t.Parallel()

	secret := "secret"
	token := rpc.NewJWTAuth(secret)

	t.Run("New", func(t *testing.T) {
		_, err := token.New()
		require.NoError(t, err)
	})

	t.Run("Verify", func(t *testing.T) {
		t.Run("good", func(t *testing.T) {
			signed, err := token.New()
			require.NoError(t, err)

			err = token.Verify(signed)
			assert.NoError(t, err)
		})

		t.Run("bad", func(t *testing.T) {
			t.Run("nonsense", func(t *testing.T) {
				err := token.Verify("blah")
				assert.Error(t, err)
			})

			t.Run("bad key", func(t *testing.T) {
				badSecret := "bad secret"
				require.NotEqual(t, badSecret, secret, "can't testing the wrong secret, we have the same one")

				bad, err := rpc.NewJWTAuth(badSecret).New()
				require.NoError(t, err)

				err = token.Verify(bad)
				assert.Error(t, err)
			})

			t.Run("wrong hashing algorithm", func(t *testing.T) {
				method := "HS256"
				require.NotEqual(t, method, rpc.JWTAlg, "can't test the wrong alg, we have the same one")

				badTok := jwt.NewWithClaims(jwt.GetSigningMethod(method), jwt.StandardClaims{})
				bad, err := badTok.SignedString([]byte(secret))
				require.NoError(t, err)

				err = token.Verify(bad)
				if assert.Error(t, err) {
					assert.EqualError(t, err, "unexpected signing method: "+method)
				}
			})

			t.Run("missing issued at", func(t *testing.T) {
				badTok := jwt.NewWithClaims(
					jwt.GetSigningMethod(rpc.JWTAlg),
					jwt.StandardClaims{
						ExpiresAt: time.Now().Unix(),
					},
				)
				bad, err := badTok.SignedString([]byte(secret))
				require.NoError(t, err)

				err = token.Verify(bad)
				if assert.Error(t, err) {
					assert.EqualError(t, err, "issued at was invalid")
				}
			})

			t.Run("missing expires at", func(t *testing.T) {
				badTok := jwt.NewWithClaims(
					jwt.GetSigningMethod(rpc.JWTAlg),
					jwt.StandardClaims{
						IssuedAt: time.Now().Unix(),
					},
				)
				bad, err := badTok.SignedString([]byte(secret))
				require.NoError(t, err)

				err = token.Verify(bad)
				if assert.Error(t, err) {
					assert.EqualError(t, err, "expires at was invalid")
				}
			})

			t.Run("lifetime is too large", func(t *testing.T) {
				badTok := jwt.NewWithClaims(
					jwt.GetSigningMethod(rpc.JWTAlg),
					jwt.StandardClaims{
						IssuedAt:  time.Now().Unix(),
						ExpiresAt: time.Now().Add(2 * rpc.JWTLifetime).Unix(),
					},
				)
				bad, err := badTok.SignedString([]byte(secret))
				require.NoError(t, err)

				err = token.Verify(bad)
				if assert.Error(t, err) {
					assert.EqualError(t, err, fmt.Sprintf("lifetime too large. Expected %s, was %s", rpc.JWTLifetime, 2*rpc.JWTLifetime))
				}
			})
		})
	})

	t.Run("GetRequestMetadata", func(t *testing.T) {
		headers, err := token.GetRequestMetadata(context.Background())

		assert.Nil(t, err)

		require.Contains(t, headers, "authorization")
		require.True(t, strings.HasPrefix(headers["authorization"], "BEARER "))

		err = token.Verify(strings.TrimLeft(headers["authorization"], "BEARER "))
		assert.NoError(t, err)
	})

	t.Run("RequireTransportSecurity", func(t *testing.T) {
		assert.False(t, token.RequireTransportSecurity())
	})
}
