package service

import (
	"context"
	"encoding/base64"
	"golang.org/x/crypto/bcrypt"
	"testing"

	genusers "github.com/Vidalee/FishyKeys/gen/users"
	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/Vidalee/FishyKeys/internal/testutil"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUsersTestService(t *testing.T) *UsersService {
	keyManager := crypto.GetDefaultKeyManager()
	masterKey, err := crypto.GenerateSecret()
	require.NoError(t, err, "failed to generate master key")
	err = keyManager.SetNewMasterKey(masterKey, 1, 1)
	require.NoError(t, err, "failed to set new master key")
	repo := repository.NewUsersRepository(testDB)
	return NewUsersService(keyManager, repo)
}

func TestUsersService_CreateUser(t *testing.T) {
	service := setupUsersTestService(t)
	ctx := context.Background()

	tests := []struct {
		name              string
		username          string
		password          string
		expectedError     bool
		expectedErrorText string
		doDuplicate       bool
	}{
		{
			name:              "empty username",
			username:          "",
			password:          "password",
			expectedError:     true,
			expectedErrorText: "username and password must be provided",
		},
		{
			name:              "password too short",
			username:          "username",
			password:          "less8",
			expectedError:     true,
			expectedErrorText: "password must be at least 8 characters long",
		},
		{
			name:          "valid user",
			username:      "username",
			password:      "long_password",
			expectedError: false,
		},
		{
			name:              "duplicate username",
			username:          "username",
			password:          "another_password",
			expectedError:     true,
			expectedErrorText: "username already exists",
			doDuplicate:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testutil.ClearTable(ctx, "users")
			require.NoError(t, err)

			payload := &genusers.CreateUserPayload{
				Username: tt.username,
				Password: tt.password,
			}

			result, err := service.CreateUser(ctx, payload)
			if tt.doDuplicate {
				result, err = service.CreateUser(ctx, payload)
			}

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedErrorText, err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Username)
				assert.Equal(t, tt.username, *result.Username)

				user, err := service.usersRepository.GetUserByUsername(ctx, tt.username)
				assert.NoError(t, err)

				assert.Equal(t, tt.username, user.Username)

				//use bcrypt
				encryptedPassword, err := base64.StdEncoding.DecodeString(user.Password)
				assert.NoError(t, err, "failed to decode encrypted password")

				err = bcrypt.CompareHashAndPassword(encryptedPassword, []byte(tt.password))
				assert.NoError(t, err, "password does not match")

				assert.NoError(t, err)
			}
		})
	}
}

func TestUsersService_AuthUser(t *testing.T) {
	service := setupUsersTestService(t)
	ctx := context.Background()

	tests := []struct {
		name              string
		createUser        bool
		username          string
		password          string
		passwordToEnter   string
		expectedError     bool
		expectedErrorText string
	}{
		{
			name:              "empty password",
			createUser:        false,
			username:          "username",
			password:          "",
			expectedError:     true,
			expectedErrorText: "username and password must be provided",
		},
		{
			name:            "correct credentials",
			createUser:      true,
			username:        "username",
			password:        "password",
			passwordToEnter: "password",
			expectedError:   false,
		},
		{
			name:              "wrong password",
			createUser:        true,
			username:          "username",
			password:          "password",
			passwordToEnter:   "wrong_password",
			expectedError:     true,
			expectedErrorText: "invalid username or password",
		},
		{
			name:              "nonexistent user",
			createUser:        false,
			username:          "nonexistent",
			password:          "password",
			passwordToEnter:   "password",
			expectedError:     true,
			expectedErrorText: "invalid username or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testutil.ClearTable(ctx, "users")
			require.NoError(t, err)

			if tt.createUser {
				createPayload := &genusers.CreateUserPayload{
					Username: tt.username,
					Password: tt.password,
				}
				_, err := service.CreateUser(ctx, createPayload)
				require.NoError(t, err, "failed to create user for auth test")
			}

			payload := &genusers.AuthUserPayload{
				Username: tt.username,
				Password: tt.passwordToEnter,
			}

			result, err := service.AuthUser(ctx, payload)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, tt.expectedErrorText, err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Username)
				assert.Equal(t, tt.username, *result.Username)

				//check token once we properly implement it
			}
		})
	}
}

func TestUsersService_ListUsers(t *testing.T) {
	service := setupUsersTestService(t)
	ctx := context.Background()
	err := testutil.ClearTable(ctx, "users")

	users, err := service.ListUsers(ctx)
	assert.NoError(t, err)
	assert.Empty(t, users, "expected no users after clearing table")

	testUsers := []genusers.CreateUserPayload{
		{Username: "user1", Password: "password1"},
		{Username: "user2", Password: "password2"},
	}

	for _, user := range testUsers {
		_, err := service.CreateUser(ctx, &user)
		assert.NoError(t, err, "failed to create test user")
	}

	users, err = service.ListUsers(ctx)
	assert.NoError(t, err, "failed to list users")
	assert.Len(t, users, len(testUsers), "expected to find all test users")
	for i, user := range users {
		assert.Equal(t, testUsers[i].Username, user.Username, "usernames should match")
	}
}
