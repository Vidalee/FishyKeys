package service

import (
	"context"
	"golang.org/x/crypto/bcrypt"
	"strconv"
	"testing"

	genkey "github.com/Vidalee/FishyKeys/gen/key_management"
	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/Vidalee/FishyKeys/internal/testutil"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupKeyTestService() *KeyManagementService {
	keyManager := crypto.GetDefaultKeyManager()
	settingsRepository := repository.NewGlobalSettingsRepository(testDB)
	usersRepository := repository.NewUsersRepository(testDB)
	rolesRepository := repository.NewRolesRepository(testDB)
	userRolesRepository := repository.NewUserRolesRepository(testDB)
	secretsRepository := repository.NewSecretsRepository(testDB)
	return NewKeyManagementService(keyManager, settingsRepository, usersRepository, rolesRepository, userRolesRepository, secretsRepository)
}

func setupKeyTestServiceWithKeyManager(keyManager *crypto.KeyManager) *KeyManagementService {
	settingsRepository := repository.NewGlobalSettingsRepository(testDB)
	usersRepository := repository.NewUsersRepository(testDB)
	rolesRepository := repository.NewRolesRepository(testDB)
	userRolesRepository := repository.NewUserRolesRepository(testDB)
	secretsRepository := repository.NewSecretsRepository(testDB)
	return NewKeyManagementService(keyManager, settingsRepository, usersRepository, rolesRepository, userRolesRepository, secretsRepository)
}

func clearKeyServiceTables(t *testing.T, ctx context.Context) {
	err := testutil.ClearTable(ctx, "global_settings")
	require.NoError(t, err)
	err = testutil.ClearTable(ctx, "users")
	require.NoError(t, err)
	err = testutil.ClearTable(ctx, "secrets")
	require.NoError(t, err)
}

func TestKeyManagementService_CreateMasterKey(t *testing.T) {
	service := setupKeyTestService()
	ctx := context.Background()

	tests := []struct {
		name          string
		totalShares   int
		minShares     int
		expectedError bool
		adminUsername string
		adminPassword string
	}{
		{
			name:          "valid parameters",
			totalShares:   5,
			minShares:     3,
			expectedError: false,
			adminUsername: "admin",
			adminPassword: "admin_password",
		},
		{
			name:          "invalid min shares",
			totalShares:   5,
			minShares:     0,
			expectedError: true,
			adminUsername: "admin",
			adminPassword: "admin_password",
		},
		{
			name:          "invalid total shares",
			totalShares:   0,
			minShares:     3,
			expectedError: true,
			adminUsername: "admin",
			adminPassword: "admin_password",
		},
		{
			name:          "min shares greater than total",
			totalShares:   3,
			minShares:     5,
			expectedError: true,
			adminUsername: "admin",
			adminPassword: "admin_password",
		},
		{
			name:          "empty admin username",
			totalShares:   3,
			minShares:     5,
			expectedError: true,
			adminUsername: "",
			adminPassword: "admin_password",
		},
		{
			name:          "admin password too short",
			totalShares:   3,
			minShares:     5,
			expectedError: true,
			adminUsername: "admin_username",
			adminPassword: "admin_password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearKeyServiceTables(t, ctx)

			payload := &genkey.CreateMasterKeyPayload{
				TotalShares:   tt.totalShares,
				MinShares:     tt.minShares,
				AdminUsername: tt.adminUsername,
				AdminPassword: tt.adminPassword,
			}

			result, err := service.CreateMasterKey(ctx, payload)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)

				settings, err := service.settingsRepository.GetSettings(ctx, columnTotalSharesColumn, columnMinSharesColumn, columnMasterKeyChecksumColumn)
				assert.Error(t, err)
				assert.Nil(t, settings)
				assert.Equal(t, "setting not found", err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Shares, tt.totalShares)

				// Verify database values
				settings, err := service.settingsRepository.GetSettings(ctx, columnTotalSharesColumn, columnMinSharesColumn, columnMasterKeyChecksumColumn)
				require.NoError(t, err)

				storedTotalShares, err := strconv.Atoi(settings[columnTotalSharesColumn])
				require.NoError(t, err)
				assert.Equal(t, tt.totalShares, storedTotalShares)

				storedMinShares, err := strconv.Atoi(settings[columnMinSharesColumn])
				require.NoError(t, err)
				assert.Equal(t, tt.minShares, storedMinShares)

				// Verify checksum exists and is not empty
				assert.NotEmpty(t, settings[columnMasterKeyChecksumColumn])

				//jwtSigningKey, exists := settings[jwtSigningKeyColumn]
				//assert.True(t, exists)
				//assert.NotEmpty(t, jwtSigningKey)
				//_, err = base64.StdEncoding.DecodeString(settings[columnMasterKeyChecksumColumn])
				//assert.NoError(t, err)

				// Verify admin user
				user, err := service.usersRepository.GetUserByUsername(ctx, tt.adminUsername)
				require.NoError(t, err)

				assert.Equal(t, tt.adminUsername, user.Username)

				err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(tt.adminPassword))
				assert.NoError(t, err)
			}
		})
	}
}

func TestKeyManagementService_GetKeyStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("no key set", func(t *testing.T) {
		clearKeyServiceTables(t, ctx)
		service := setupKeyTestService()
		result, err := service.GetKeyStatus(ctx)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("key exists and unlocked", func(t *testing.T) {
		clearKeyServiceTables(t, ctx)
		service := setupKeyTestService()

		createPayload := &genkey.CreateMasterKeyPayload{
			TotalShares:   5,
			MinShares:     3,
			AdminUsername: "admin",
			AdminPassword: "admin_password",
		}
		_, err := service.CreateMasterKey(ctx, createPayload)
		require.NoError(t, err)

		result, err := service.GetKeyStatus(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsLocked)
		assert.Equal(t, 3, result.MinShares)
		assert.Equal(t, 5, result.TotalShares)
		assert.Equal(t, 0, result.CurrentShares)
	})

	t.Run("key exists but locked", func(t *testing.T) {
		clearKeyServiceTables(t, ctx)
		setupService := setupKeyTestService()
		createPayload := &genkey.CreateMasterKeyPayload{
			TotalShares:   5,
			MinShares:     3,
			AdminUsername: "admin",
			AdminPassword: "admin_password",
		}
		_, err := setupService.CreateMasterKey(ctx, createPayload)
		require.NoError(t, err)

		// Create a new service instance to test locked state
		newService := setupKeyTestService()
		result, err := newService.GetKeyStatus(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsLocked)
		assert.Equal(t, 3, result.MinShares)
		assert.Equal(t, 5, result.TotalShares)
		assert.Equal(t, 0, result.CurrentShares)
	})
}

func TestKeyManagementService_AddShare(t *testing.T) {
	ctx := context.Background()

	t.Run("no key set", func(t *testing.T) {
		clearKeyServiceTables(t, ctx)
		service := setupKeyTestService()
		payload := &genkey.AddSharePayload{
			Share: "invalid_share",
		}

		result, err := service.AddShare(ctx, payload)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("valid shares sequence", func(t *testing.T) {
		clearKeyServiceTables(t, ctx)
		setupService := setupKeyTestService()

		createPayload := &genkey.CreateMasterKeyPayload{
			TotalShares:   5,
			MinShares:     3,
			AdminUsername: "admin",
			AdminPassword: "admin_password",
		}
		createResult, err := setupService.CreateMasterKey(ctx, createPayload)
		require.NoError(t, err)

		service := setupKeyTestService()

		for i := 0; i < 3; i++ {
			payload := &genkey.AddSharePayload{
				Share: createResult.Shares[i],
			}

			result, err := service.AddShare(ctx, payload)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, i, result.Index)

			status, err := service.GetKeyStatus(ctx)
			require.NoError(t, err)
			assert.Equal(t, i+1, status.CurrentShares)

			// Should be unlocked only after adding minimum shares
			if i < 2 {
				assert.True(t, status.IsLocked)
				assert.False(t, result.Unlocked)
			} else {
				assert.False(t, status.IsLocked)
				assert.True(t, result.Unlocked)
			}
		}

		// Try adding more shares after unlocked
		payload := &genkey.AddSharePayload{
			Share: createResult.Shares[3],
		}
		result, err := service.AddShare(ctx, payload)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("valid shares but checksum wrong", func(t *testing.T) {
		clearKeyServiceTables(t, ctx)
		setupService := setupKeyTestService()

		createPayload := &genkey.CreateMasterKeyPayload{
			TotalShares:   5,
			MinShares:     3,
			AdminUsername: "admin",
			AdminPassword: "admin_password",
		}
		createResult, err := setupService.CreateMasterKey(ctx, createPayload)
		require.NoError(t, err)

		service := setupKeyTestService()

		// Tamper with the checksum in the DB
		err = service.settingsRepository.StoreSetting(ctx, columnMasterKeyChecksumColumn, "d3adbeef")
		require.NoError(t, err)

		for i := range 3 {
			payload := &genkey.AddSharePayload{
				Share: createResult.Shares[i],
			}

			result, err := service.AddShare(ctx, payload)
			if i < 2 {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, i, result.Index)
				assert.False(t, result.Unlocked)
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Equal(t, "The key recombined from the shares is not the correct key", err.Error())
			}

			status, err := service.GetKeyStatus(ctx)
			require.NoError(t, err)
			assert.True(t, status.IsLocked)
		}
	})

	t.Run("invalid share", func(t *testing.T) {
		clearKeyServiceTables(t, ctx)
		setupService := setupKeyTestService()

		createPayload := &genkey.CreateMasterKeyPayload{
			TotalShares:   5,
			MinShares:     3,
			AdminUsername: "admin",
			AdminPassword: "admin_password",
		}
		createResult, err := setupService.CreateMasterKey(ctx, createPayload)
		require.NoError(t, err)

		service := setupKeyTestService()

		for i := 0; i < 2; i++ {
			payload := &genkey.AddSharePayload{
				Share: createResult.Shares[i],
			}
			_, err = service.AddShare(ctx, payload)
			require.NoError(t, err)
		}

		payload := &genkey.AddSharePayload{
			Share: "invalid_share",
		}
		result, err := service.AddShare(ctx, payload)
		assert.Error(t, err)
		assert.Nil(t, result)

		status, err := service.GetKeyStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, status.CurrentShares)
		assert.True(t, status.IsLocked)
	})
}

func TestKeyManagementService_DeleteShare(t *testing.T) {
	ctx := context.Background()

	t.Run("no key set", func(t *testing.T) {
		clearKeyServiceTables(t, ctx)
		service := setupKeyTestService()
		payload := &genkey.DeleteSharePayload{
			Index: 0,
		}

		err := service.DeleteShare(ctx, payload)
		assert.Error(t, err)
	})

	t.Run("invalid index", func(t *testing.T) {
		clearKeyServiceTables(t, ctx)
		setupService := setupKeyTestService()

		createPayload := &genkey.CreateMasterKeyPayload{
			TotalShares:   5,
			MinShares:     3,
			AdminUsername: "admin",
			AdminPassword: "admin_password",
		}
		createResult, err := setupService.CreateMasterKey(ctx, createPayload)
		require.NoError(t, err)

		service := setupKeyTestService()

		for i := 0; i < 2; i++ {
			addPayload := &genkey.AddSharePayload{
				Share: createResult.Shares[i],
			}
			_, err = service.AddShare(ctx, addPayload)
			require.NoError(t, err)
		}

		payload := &genkey.DeleteSharePayload{
			Index: 10,
		}

		err = service.DeleteShare(ctx, payload)
		assert.Error(t, err)

		status, err := service.GetKeyStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, status.CurrentShares)
	})

	t.Run("valid share deletion", func(t *testing.T) {
		clearKeyServiceTables(t, ctx)
		setupService := setupKeyTestService()

		createPayload := &genkey.CreateMasterKeyPayload{
			TotalShares:   5,
			MinShares:     3,
			AdminUsername: "admin",
			AdminPassword: "admin_password",
		}
		createResult, err := setupService.CreateMasterKey(ctx, createPayload)
		require.NoError(t, err)

		service := setupKeyTestService()

		for i := 0; i < 2; i++ {
			addPayload := &genkey.AddSharePayload{
				Share: createResult.Shares[i],
			}
			_, err = service.AddShare(ctx, addPayload)
			require.NoError(t, err)
		}

		status, err := service.GetKeyStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, status.CurrentShares)
		assert.True(t, status.IsLocked)

		payload := &genkey.DeleteSharePayload{
			Index: 1,
		}
		err = service.DeleteShare(ctx, payload)
		assert.NoError(t, err)

		status, err = service.GetKeyStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, status.CurrentShares)
		assert.True(t, status.IsLocked)
	})
}

func TestKeyManagementService_EndToEndUnlock(t *testing.T) {
	ctx := context.Background()
	clearKeyServiceTables(t, ctx)
	setupService := setupKeyTestService()

	// Step 1: Create master key with 5 total, 3 min shares
	createPayload := &genkey.CreateMasterKeyPayload{
		TotalShares:   5,
		MinShares:     3,
		AdminUsername: "admin",
		AdminPassword: "admin_password",
	}
	createResult, err := setupService.CreateMasterKey(ctx, createPayload)
	require.NoError(t, err)
	require.Len(t, createResult.Shares, 5)

	service := setupKeyTestService()

	// Step 2: Add a wrong share (mutate one letter from a correct share)
	wrongShareBytes := []byte(createResult.Shares[0])
	if wrongShareBytes[len(wrongShareBytes)-1] != 'A' {
		wrongShareBytes[len(wrongShareBytes)-1] = 'A'
	} else {
		wrongShareBytes[len(wrongShareBytes)-1] = 'B'
	}
	wrongShare := string(wrongShareBytes)
	addPayloadWrong := &genkey.AddSharePayload{Share: wrongShare}
	addResultWrong, err := service.AddShare(ctx, addPayloadWrong)
	assert.NoError(t, err)
	assert.NotNil(t, addResultWrong)
	assert.False(t, addResultWrong.Unlocked)

	// Step 3: Add a good share (index 1)
	addPayload1 := &genkey.AddSharePayload{Share: createResult.Shares[1]}
	addResult1, err := service.AddShare(ctx, addPayload1)
	assert.NoError(t, err)
	assert.NotNil(t, addResult1)
	assert.False(t, addResult1.Unlocked)

	// Step 4: Remove share at index 0 (the wrong one)
	deletePayload := &genkey.DeleteSharePayload{Index: 0}
	err = service.DeleteShare(ctx, deletePayload)
	assert.NoError(t, err)

	// Step 5: Add two more good shares (index 2 and 3)
	addPayload2 := &genkey.AddSharePayload{Share: createResult.Shares[2]}
	addResult2, err := service.AddShare(ctx, addPayload2)
	assert.NoError(t, err)
	assert.NotNil(t, addResult2)
	assert.False(t, addResult2.Unlocked)

	addPayload3 := &genkey.AddSharePayload{Share: createResult.Shares[3]}
	addResult3, err := service.AddShare(ctx, addPayload3)
	assert.NoError(t, err)
	assert.NotNil(t, addResult3)
	// Should unlock now
	assert.True(t, addResult3.Unlocked)

	// Step 6: Check if unlocked
	finalStatus, err := service.GetKeyStatus(ctx)
	assert.NoError(t, err)
	assert.False(t, finalStatus.IsLocked)
	assert.Equal(t, 3, finalStatus.MinShares)
	assert.Equal(t, 5, finalStatus.TotalShares)
	assert.Equal(t, 3, finalStatus.CurrentShares)
}
