package authorization_test

import (
	"booker/internal/lib/api/authorization"
	"log/slog"
	"strings"
	"testing"
)

func TestAuthorization(t *testing.T) {
	tests := []struct {
		TestName      string
		Token         string
		secretKey     string
		errorExpected bool
		errorContains string
	}{
		{"valid token and secret", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiaWF0IjoxNzMzNjAzOTMxLCJleHAiOjE4NDkxODc5MzF9.127K4MnLBcYL78KKrNmzlnDLTlWIwawv5MecxayDkds", "256bitsvalid256bitsvalid256bitsvalid", false, ""},
		{"random string in token field token", "invalid", "256bitsvalid256bitsvalid256bitsvalid", true, "failed to parse token: token is malformed: token contains an invalid number of segments"},
		{"token with invalid sign", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiaWF0IjoxNzMzNjAzOTMxLCJleHAiOjE4NDkxODc5MzF9.veHC9BFqcSFow5gOyTlalESHRKgdFEnn23FdFB3Z7dw", "256bitsvalid256bitsvalid256bitsvalid", true, "failed to parse token: token signature is invalid: signature is invalid"},
		{"expired token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiaWF0IjoxNzMzNjAzOTMxLCJleHAiOjE3MzM2MDM5MzF9.8m0HyNu-MON179DukOYCRHnlDriTU7y_4qhdor8DZJ8", "256bitsvalid256bitsvalid256bitsvalid", true, "failed to parse token: token has invalid claims: token is expired"},
	}

	for _, test := range tests {
		t.Run(test.TestName, func(t *testing.T) {
			log := slog.Default()
			claims, err := authorization.Authorization(test.Token, test.secretKey)
			if test.errorExpected {
				if err == nil {
					t.Error("Authorization is not failed as expected. Error: ", err)
				}
				if !strings.Contains(err.Error(), test.errorContains) {
					log.Error("Authorization error not contain expected text.", "Error: ", err)
					t.Error("Authorization error is not correspond expected error.", "Error: ", err)
				}
				return
			}
			if err != nil {
				t.Error("Authorization is failed. Error: ", err)
			} else {
				log.Info("Authorized token: ", "claims", claims)
			}

		})
	}
}
