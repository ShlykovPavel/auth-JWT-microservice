package users_tests

import (
	"log"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashUserPassword(t *testing.T) {

	tests := []struct {
		TestName        string
		Password        string
		ComparePassword string

		ErrorExpected bool
		ErrorContains string
	}{
		{Password: "password", ComparePassword: "password", TestName: "Test valid password", ErrorExpected: false, ErrorContains: ""},
		{Password: "test", ComparePassword: "wrong", TestName: "Test wrong password in compare", ErrorExpected: true, ErrorContains: "hashedPassword is not the hash of the given password"},
		{Password: "", ComparePassword: "", TestName: "Test empty password", ErrorExpected: false, ErrorContains: ""},
		{TestName: "Empty compare password", Password: "test", ComparePassword: "", ErrorExpected: true, ErrorContains: "hashedPassword is not the hash of the given password"},
	}
	for _, tt := range tests {
		t.Run(tt.TestName, func(t *testing.T) {
			passwordHash, err := bcrypt.GenerateFromPassword([]byte(tt.Password), bcrypt.DefaultCost)
			if err != nil {
				t.Error("Hashing password is failed. Error: ", err.Error())

			}
			if passwordHash == nil {
				t.Error("Hashing password return nil")
			}
			log.Print("Password hash: ", string(passwordHash))
			err = bcrypt.CompareHashAndPassword(passwordHash, []byte(tt.ComparePassword))
			if tt.ErrorExpected {
				if err == nil {
					t.Error("CompareHashAndPassword is not failed. Error: ", err)
				}
				if !strings.Contains(err.Error(), tt.ErrorContains) {
					t.Error("CompareHashAndPassword error not contain expected text. Error: ", err)

				}
				log.Print("CompareHashAndPassword expected error: ", err)
				return
			}
			if err != nil {
				t.Error("CompareHashAndPassword is failed. Error: ", err.Error())
			} else {
				log.Print("password compared successfully")
			}

		})
	}
}

func TestSaltUserPassword(t *testing.T) {
	tests := []struct {
		TestName       string
		Password       string
		SecondPassword string
	}{
		{TestName: "2 same passwords hash are different", Password: "test", SecondPassword: "test"},
	}
	for _, tt := range tests {
		t.Run(tt.TestName, func(t *testing.T) {
			passwordHash1, err := bcrypt.GenerateFromPassword([]byte(tt.Password), bcrypt.DefaultCost)
			if err != nil {
				t.Fatal("Hashing password is failed. Error: ", err)
			}
			if passwordHash1 == nil {
				t.Fatal("Hashing password return nil")
			}
			log.Print("Password hash: ", string(passwordHash1))
			passwordHash2, err := bcrypt.GenerateFromPassword([]byte(tt.SecondPassword), bcrypt.DefaultCost)
			if err != nil {
				t.Fatal("Hashing password 2 is failed. Error: ", err)
			}
			if passwordHash2 == nil {
				t.Fatal("Hashing password 2 return nil")
			}
			log.Print("Password hash: ", string(passwordHash2))
			if string(passwordHash1) == string(passwordHash2) {
				t.Fatal("Both password hashes are equal")
			}
		})
	}
}
