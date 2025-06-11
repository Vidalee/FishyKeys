package service

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	genkey "github.com/Vidalee/FishyKeys/gen/fishykeys"
	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/Vidalee/FishyKeys/internal/testutil"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testDB *pgxpool.Pool

func setupTestService() *KeyManagementService {
	crypto.ResetKeyManager()
	keyManager := crypto.GetKeyManager()
	repo := repository.NewGlobalSettingsRepository(testDB)
	return NewKeyManagementService(keyManager, repo)
}

func TestMain(m *testing.M) {
	var err error
	testDB, err = testutil.SetupTestDB()
	if err != nil {
		panic(fmt.Sprintf("failed to setup test db: %v", err))
	}
	defer func() {
		if err := testutil.TeardownTestDB(); err != nil {
			panic(fmt.Sprintf("failed to teardown test db: %v", err))
		}
	}()

	m.Run()
}

func TestKeyManagementService_CreateMasterKey(t *testing.T) {
	service := setupTestService()
	ctx := context.Background()

	tests := []struct {
		name          string
		totalShares   int
		minShares     int
		expectedError bool
	}{
		{
			name:          "valid parameters",
			totalShares:   5,
			minShares:     3,
			expectedError: false,
		},
		{
			name:          "invalid min shares",
			totalShares:   5,
			minShares:     0,
			expectedError: true,
		},
		{
			name:          "invalid total shares",
			totalShares:   0,
			minShares:     3,
			expectedError: true,
		},
		{
			name:          "min shares greater than total",
			totalShares:   3,
			minShares:     5,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := testutil.ClearTable(ctx, "global_settings")
			require.NoError(t, err)

			payload := &genkey.CreateMasterKeyPayload{
				TotalShares: tt.totalShares,
				MinShares:   tt.minShares,
			}

			result, err := service.CreateMasterKey(ctx, payload)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result.Shares, tt.totalShares)

				// Verify database values
				settings, err := service.settingsRepository.GetSettings(ctx, columnTotalShares, columnMinShares, columnMasterKeyChecksum)
				require.NoError(t, err)

				storedTotalShares, err := strconv.Atoi(settings[columnTotalShares])
				require.NoError(t, err)
				assert.Equal(t, tt.totalShares, storedTotalShares)

				storedMinShares, err := strconv.Atoi(settings[columnMinShares])
				require.NoError(t, err)
				assert.Equal(t, tt.minShares, storedMinShares)

				// Verify checksum exists and is not empty
				assert.NotEmpty(t, settings[columnMasterKeyChecksum])
			}
		})
	}
}

func TestKeyManagementService_GetKeyStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("no key set", func(t *testing.T) {
		err := testutil.ClearTable(ctx, "global_settings")
		require.NoError(t, err)
		service := setupTestService()
		result, err := service.GetKeyStatus(ctx)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("key exists and unlocked", func(t *testing.T) {
		err := testutil.ClearTable(ctx, "global_settings")
		require.NoError(t, err)
		service := setupTestService()

		// First create a master key
		payload := &genkey.CreateMasterKeyPayload{
			TotalShares: 5,
			MinShares:   3,
		}
		_, err = service.CreateMasterKey(ctx, payload)
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
		err := testutil.ClearTable(ctx, "global_settings")
		require.NoError(t, err)
		// First create a service and set up the key
		setupService := setupTestService()
		payload := &genkey.CreateMasterKeyPayload{
			TotalShares: 5,
			MinShares:   3,
		}
		_, err = setupService.CreateMasterKey(ctx, payload)
		require.NoError(t, err)

		// Create a new service instance to test locked state
		newService := setupTestService()
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
		err := testutil.ClearTable(ctx, "global_settings")
		require.NoError(t, err)
		service := setupTestService()
		payload := &genkey.AddSharePayload{
			Share: "invalid_share",
		}

		result, err := service.AddShare(ctx, payload)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("valid shares sequence", func(t *testing.T) {
		err := testutil.ClearTable(ctx, "global_settings")
		require.NoError(t, err)
		setupService := setupTestService()

		// First create a master key
		createPayload := &genkey.CreateMasterKeyPayload{
			TotalShares: 5,
			MinShares:   3,
		}
		createResult, err := setupService.CreateMasterKey(ctx, createPayload)
		require.NoError(t, err)

		service := setupTestService()

		// Add shares one by one and verify status
		for i := 0; i < 3; i++ {
			payload := &genkey.AddSharePayload{
				Share: createResult.Shares[i],
			}

			result, err := service.AddShare(ctx, payload)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, i, result.Index)

			// Verify key status after each share
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

	t.Run("invalid share", func(t *testing.T) {
		err := testutil.ClearTable(ctx, "global_settings")
		require.NoError(t, err)
		setupService := setupTestService()

		// First create a master key
		createPayload := &genkey.CreateMasterKeyPayload{
			TotalShares: 5,
			MinShares:   3,
		}
		createResult, err := setupService.CreateMasterKey(ctx, createPayload)
		require.NoError(t, err)

		service := setupTestService()

		// Add minimum shares first
		for i := 0; i < 2; i++ {
			payload := &genkey.AddSharePayload{
				Share: createResult.Shares[i],
			}
			_, err = service.AddShare(ctx, payload)
			require.NoError(t, err)
		}

		// Try adding an invalid share
		payload := &genkey.AddSharePayload{
			Share: "invalid_share",
		}
		result, err := service.AddShare(ctx, payload)
		assert.Error(t, err)
		assert.Nil(t, result)

		// Verify status hasn't changed
		status, err := service.GetKeyStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, status.CurrentShares)
		assert.True(t, status.IsLocked)
	})
}

func TestKeyManagementService_DeleteShare(t *testing.T) {
	ctx := context.Background()

	t.Run("no key set", func(t *testing.T) {
		err := testutil.ClearTable(ctx, "global_settings")
		require.NoError(t, err)
		service := setupTestService()
		payload := &genkey.DeleteSharePayload{
			Index: 0,
		}

		err = service.DeleteShare(ctx, payload)
		assert.Error(t, err)
	})

	t.Run("invalid index", func(t *testing.T) {
		err := testutil.ClearTable(ctx, "global_settings")
		require.NoError(t, err)
		setupService := setupTestService()

		// First create a master key
		createPayload := &genkey.CreateMasterKeyPayload{
			TotalShares: 5,
			MinShares:   3,
		}
		createResult, err := setupService.CreateMasterKey(ctx, createPayload)
		require.NoError(t, err)

		service := setupTestService()

		// Add some shares
		for i := 0; i < 2; i++ {
			addPayload := &genkey.AddSharePayload{
				Share: createResult.Shares[i],
			}
			_, err = service.AddShare(ctx, addPayload)
			require.NoError(t, err)
		}

		// Try deleting an invalid index
		payload := &genkey.DeleteSharePayload{
			Index: 10, // Invalid index
		}

		err = service.DeleteShare(ctx, payload)
		assert.Error(t, err)

		// Verify status hasn't changed
		status, err := service.GetKeyStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, status.CurrentShares)
	})

	t.Run("valid share deletion", func(t *testing.T) {
		err := testutil.ClearTable(ctx, "global_settings")
		require.NoError(t, err)
		setupService := setupTestService()

		// First create a master key
		createPayload := &genkey.CreateMasterKeyPayload{
			TotalShares: 5,
			MinShares:   3,
		}
		createResult, err := setupService.CreateMasterKey(ctx, createPayload)
		require.NoError(t, err)

		service := setupTestService()

		// Add some shares
		for i := 0; i < 2; i++ {
			addPayload := &genkey.AddSharePayload{
				Share: createResult.Shares[i],
			}
			_, err = service.AddShare(ctx, addPayload)
			require.NoError(t, err)
		}

		// Verify initial status
		status, err := service.GetKeyStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, status.CurrentShares)
		assert.True(t, status.IsLocked)

		// Delete a share
		payload := &genkey.DeleteSharePayload{
			Index: 1,
		}
		err = service.DeleteShare(ctx, payload)
		assert.NoError(t, err)

		// Verify status after deletion
		status, err = service.GetKeyStatus(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, status.CurrentShares)
		assert.True(t, status.IsLocked)
	})
}
