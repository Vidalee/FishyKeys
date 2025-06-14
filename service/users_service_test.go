package service

import (
	"context"
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
				require.NoError(t, err)

				assert.Equal(t, tt.username, user.Username)

				decryptedPassword, err := crypto.Decrypt(service.keyManager, user.Password)
				require.NoError(t, err)
				assert.Equal(t, tt.password, string(decryptedPassword))
			}
		})
	}
}
