package service

import (
	"context"
	"encoding/base64"
	"fmt"
	genkey "github.com/Vidalee/FishyKeys/gen/key_management"
	"github.com/golang-jwt/jwt/v5"
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
	usersRepo := repository.NewUsersRepository(testDB)
	globalSettingsRepo := repository.NewGlobalSettingsRepository(testDB)
	secretsRepo := repository.NewSecretsRepository(testDB)
	return NewUsersService(keyManager, usersRepo, globalSettingsRepo, secretsRepo)
}

func clearUsersServiceTables(t *testing.T, ctx context.Context) {
	err := testutil.ClearTable(ctx, "global_settings")
	require.NoError(t, err)
	err = testutil.ClearTable(ctx, "users")
	require.NoError(t, err)
	err = testutil.ClearTable(ctx, "secrets")
	require.NoError(t, err)
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
			clearUsersServiceTables(t, ctx)

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
				assert.NotEqual(t, 0, user.ID, "user ID should not be zero")
				err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(tt.password))
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
			name:              "system user",
			createUser:        false,
			username:          "system",
			passwordToEnter:   "deactivated_password",
			expectedError:     true,
			expectedErrorText: "system user cannot be authenticated",
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
			clearUsersServiceTables(t, ctx)

			var createUserResult *genusers.CreateUserResult
			if tt.createUser {
				createMasterKeyPayload := &genkey.CreateMasterKeyPayload{
					TotalShares:   3,
					MinShares:     2,
					AdminUsername: "admin",
					AdminPassword: "password",
				}

				keyService := setupKeyTestServiceWithKeyManager(service.keyManager)
				_, err := keyService.CreateMasterKey(ctx, createMasterKeyPayload)
				require.NoError(t, err, "failed to create master key for auth test")

				createPayload := &genusers.CreateUserPayload{
					Username: tt.username,
					Password: tt.password,
				}
				createUserResult, err = service.CreateUser(ctx, createPayload)
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

				assert.NotNil(t, result.Token, "token should not be nil")
				token, err := jwt.ParseWithClaims(*result.Token, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, jwt.ErrSignatureInvalid
					}
					decryptedSecret, err := service.secretsRepository.GetSecretByPath(ctx, service.keyManager, "internal/jwt_signing_key")
					if err != nil {
						return nil, genusers.MakeInternalError(fmt.Errorf("could not retrieve JWT signing key: %w", err))
					}

					return base64.StdEncoding.DecodeString(decryptedSecret.DecryptedValue)
				})
				assert.NoError(t, err, "failed to parse JWT token")
				assert.True(t, token.Valid, "token should be valid")
				claims, ok := token.Claims.(*JwtClaims)
				assert.True(t, ok, "claims should be of type JwtClaims")
				assert.Equal(t, tt.username, claims.Username, "username in claims should match")
				assert.Equal(t, *createUserResult.ID, claims.UserID, "user ID in token should match created user ID")
			}
		})
	}
}

func TestUsersService_ListUsers(t *testing.T) {
	service := setupUsersTestService(t)
	ctx := context.Background()

	clearUsersServiceTables(t, ctx)

	users, err := service.ListUsers(ctx)
	assert.NoError(t, err)
	// 1 since there is a system user
	assert.Len(t, users, 1, "expected no users after clearing table")

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
	// +1 since there is a system user
	assert.Len(t, users, len(testUsers)+1, "expected to find all test users")

	// Skip the system user at index 0
	for i := 1; i < len(users); i++ {
		user := users[i]
		assert.Equal(t, testUsers[i-1].Username, user.Username, "usernames should match")
	}
}

func TestUsersService_DeleteUser(t *testing.T) {
	service := setupUsersTestService(t)
	ctx := context.Background()

	tests := []struct {
		name              string
		username          string
		createUser        bool
		expectedError     bool
		expectedErrorText string
	}{
		{
			name:              "empty username",
			username:          "",
			createUser:        false,
			expectedError:     true,
			expectedErrorText: "username must be provided",
		},
		{
			name:              "nonexistent user",
			username:          "username",
			createUser:        false,
			expectedError:     true,
			expectedErrorText: "user not found",
		},
		{
			name:          "existing user",
			username:      "username",
			createUser:    true,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearUsersServiceTables(t, ctx)

			if tt.createUser {
				createPayload := &genusers.CreateUserPayload{
					Username: tt.username,
					Password: "password",
				}
				_, err := service.CreateUser(ctx, createPayload)
				require.NoError(t, err, "failed to create user for auth test")
			}

			payload := &genusers.DeleteUserPayload{
				Username: tt.username,
			}

			err := service.DeleteUser(ctx, payload)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorText, err.Error())
			} else {
				assert.NoError(t, err)

				_, err := service.usersRepository.GetUserByUsername(ctx, tt.username)
				assert.Error(t, err)
				assert.Equal(t, repository.ErrUserNotFound, err, "expected user to be deleted")
			}
		})
	}
}
