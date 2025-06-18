package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	genusers "github.com/Vidalee/FishyKeys/gen/users"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strconv"

	"github.com/Vidalee/FishyKeys/repository"

	genkey "github.com/Vidalee/FishyKeys/gen/key_management"
	"github.com/Vidalee/FishyKeys/internal/crypto"
)

var (
	ErrInvalidParameters = errors.New("invalid parameters")
)

const (
	columnTotalSharesColumn       = "total_shares"
	columnMinSharesColumn         = "min_shares"
	columnMasterKeyChecksumColumn = "master_key_checksum"
	checksumExpectedValue         = "fishykeys_checksum"
)

type KeyManagementService struct {
	keyManager          *crypto.KeyManager
	settingsRepository  repository.GlobalSettingsRepository
	usersRepository     repository.UsersRepository
	rolesRepository     repository.RolesRepository
	userRolesRepository repository.UserRolesRepository
	secretsRepository   repository.SecretsRepository
}

func NewKeyManagementService(
	keyManager *crypto.KeyManager,
	settingsRepository repository.GlobalSettingsRepository,
	usersRepository repository.UsersRepository,
	rolesRepository repository.RolesRepository,
	userRolesRepository repository.UserRolesRepository,
	secretsRepository repository.SecretsRepository,
) *KeyManagementService {
	keySettings, err := settingsRepository.GetSettings(context.Background(), columnMasterKeyChecksumColumn, columnTotalSharesColumn, columnMinSharesColumn)
	if err != nil {
		if !errors.Is(err, repository.ErrSettingNotFound) {
			log.Fatalf("error retrieving key settings on service init: %v", err)
		}
	} else {
		minShares, err := strconv.Atoi(keySettings[columnMinSharesColumn])
		if err != nil {
			log.Fatalf("error parsing min shares from db: %v", err)
		}
		totalShares, err := strconv.Atoi(keySettings[columnTotalSharesColumn])
		if err != nil {
			log.Fatalf("error parsing total shares from db: %v", err)
		}
		err = keyManager.ConfigureKeySystem(minShares, totalShares)
		if err != nil {
			log.Fatalf("error configuring key system with existing shares: %v", err)
		}
	}

	return &KeyManagementService{
		keyManager:          keyManager,
		settingsRepository:  settingsRepository,
		usersRepository:     usersRepository,
		rolesRepository:     rolesRepository,
		userRolesRepository: userRolesRepository,
		secretsRepository:   secretsRepository,
	}
}

func (s *KeyManagementService) CreateMasterKey(ctx context.Context, payload *genkey.CreateMasterKeyPayload) (*genkey.CreateMasterKeyResult, error) {
	if payload.TotalShares <= 0 || payload.MinShares <= 0 || payload.MinShares > payload.TotalShares {
		return nil, ErrInvalidParameters
	}

	if payload.AdminUsername == "" || payload.AdminPassword == "" {
		return nil, genusers.MakeInvalidParameters(fmt.Errorf("username and password must be provided"))
	}

	if len(payload.AdminPassword) < 8 {
		return nil, genusers.MakeInvalidParameters(fmt.Errorf("password must be at least 8 characters long"))
	}

	_, err := s.settingsRepository.GetSetting(ctx, columnMasterKeyChecksumColumn)
	if err == nil {
		return nil, genkey.KeyAlreadyExists("master key already exists")
	}

	masterKey, err := crypto.GenerateSecret()
	if err != nil {
		return nil, genkey.MakeInternalError(fmt.Errorf("error generating master key: %v", err))
	}
	checksum, err := crypto.EncryptWithKey(masterKey, []byte(checksumExpectedValue))
	if err != nil {
		return nil, genkey.MakeInternalError(fmt.Errorf("error encrypting master key checksum: %v", err))
	}
	shares, err := crypto.SplitSecret(masterKey, payload.TotalShares, payload.MinShares)
	if err != nil {
		return nil, genkey.MakeInternalError(fmt.Errorf("error splitting secret into shares: %v", err))
	}

	encodedShares := make([]string, len(shares))
	for i, b := range shares {
		encodedShares[i] = base64.StdEncoding.EncodeToString(b)
	}

	err = s.settingsRepository.StoreSettings(ctx, map[string]string{
		columnTotalSharesColumn:       strconv.Itoa(payload.TotalShares),
		columnMinSharesColumn:         strconv.Itoa(payload.MinShares),
		columnMasterKeyChecksumColumn: checksum,
	})
	if err != nil {
		return nil, genkey.MakeInternalError(fmt.Errorf("error storing key settings: %v", err))
	}

	err = s.keyManager.SetNewMasterKey(masterKey, payload.MinShares, payload.TotalShares)
	if err != nil {
		return nil, genkey.MakeInternalError(fmt.Errorf("error setting new master key: %v", err))
	}

	jwtSigningKey, err := crypto.GenerateSecret()
	if err != nil {
		s.keyManager.RollbackToUninitialized()
		delErr := s.settingsRepository.DeleteSettings(ctx, columnTotalSharesColumn, columnMinSharesColumn, columnMasterKeyChecksumColumn)

		if delErr != nil {
			return nil, genkey.MakeInternalError(fmt.Errorf(
				"error generating JWT signing key: %v; rollback cleanup also failed: %v", err, delErr))
		}
		return nil, genkey.MakeInternalError(fmt.Errorf("error generating JWT signing key: %v", err))
	}
	_, err = s.secretsRepository.CreateSecret(ctx, s.keyManager, "internal/jwt_signing_key", base64.StdEncoding.EncodeToString(jwtSigningKey))
	if err != nil {
		s.keyManager.RollbackToUninitialized()
		delErr := s.settingsRepository.DeleteSettings(ctx, columnTotalSharesColumn, columnMinSharesColumn, columnMasterKeyChecksumColumn)

		if delErr != nil {
			return nil, genkey.MakeInternalError(fmt.Errorf(
				"error creating JWT signing key: %v; rollback cleanup also failed: %v", err, delErr))
		}
		return nil, genkey.MakeInternalError(fmt.Errorf("error storing JWT signing key: %v", err))
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		s.keyManager.RollbackToUninitialized()
		jwtDelErr := s.secretsRepository.DeleteSecret(ctx, "internal/jwt_signing_key")
		delErr := s.settingsRepository.DeleteSettings(ctx, columnTotalSharesColumn, columnMinSharesColumn, columnMasterKeyChecksumColumn)
		if delErr != nil {
			if jwtDelErr != nil {
				return nil, genkey.MakeInternalError(fmt.Errorf(
					"could not create admin user: %v; rollback cleanup also failed: %v; jwt deletion failed: %v", err, delErr, jwtDelErr))
			}
			return nil, genkey.MakeInternalError(fmt.Errorf(
				"could not create admin user: %v; rollback cleanup also failed: %v", err, delErr))
		}

		return nil, genkey.MakeInternalError(fmt.Errorf("could not create admin user: %v", err))
	}

	userId, err := s.usersRepository.CreateUser(ctx, payload.AdminUsername, string(encryptedPassword))
	if err != nil {
		s.keyManager.RollbackToUninitialized()
		jwtDelErr := s.secretsRepository.DeleteSecret(ctx, "internal/jwt_signing_key")
		delErr := s.settingsRepository.DeleteSettings(ctx, columnTotalSharesColumn, columnMinSharesColumn, columnMasterKeyChecksumColumn)
		if delErr != nil {
			if jwtDelErr != nil {
				return nil, genkey.MakeInternalError(fmt.Errorf(
					"could not create admin user: %v; rollback cleanup also failed: %v; jwt deletion failed: %v", err, delErr, jwtDelErr))
			}
			return nil, genkey.MakeInternalError(fmt.Errorf(
				"could not create admin user: %v; rollback cleanup also failed: %v", err, delErr))
		}

		return nil, genkey.MakeInternalError(fmt.Errorf("could not create admin user: %v", err))
	}

	role, err := s.rolesRepository.GetRoleByName(ctx, "admin")
	if err != nil {
		s.keyManager.RollbackToUninitialized()
		jwtDelErr := s.secretsRepository.DeleteSecret(ctx, "internal/jwt_signing_key")
		delErr := s.settingsRepository.DeleteSettings(ctx, columnTotalSharesColumn, columnMinSharesColumn, columnMasterKeyChecksumColumn)
		if delErr != nil {
			if jwtDelErr != nil {
				return nil, genkey.MakeInternalError(fmt.Errorf(
					"could not create admin user: %v; rollback cleanup also failed: %v; jwt deletion failed: %v", err, delErr, jwtDelErr))
			}
			return nil, genkey.MakeInternalError(fmt.Errorf(
				"could not create admin user: %v; rollback cleanup also failed: %v", err, delErr))
		}

		return nil, genkey.MakeInternalError(fmt.Errorf("could not create admin user: %v", err))
	}

	err = s.userRolesRepository.AssignRoleToUser(ctx, userId, role.ID)
	if err != nil {
		return nil, genkey.MakeInternalError(fmt.Errorf("could not assign role to admin user: %v", err))
	}

	return &genkey.CreateMasterKeyResult{
		Shares:        encodedShares,
		AdminUsername: &payload.AdminUsername,
	}, nil
}

func (s *KeyManagementService) GetKeyStatus(ctx context.Context) (*genkey.GetKeyStatusResult, error) {
	_, err := s.settingsRepository.GetSetting(ctx, columnMasterKeyChecksumColumn)
	if err != nil {
		if errors.Is(err, repository.ErrSettingNotFound) {
			return nil, genkey.NoKeySet("master key not set")
		}
		return nil, genkey.InternalError("error retrieving key status: " + err.Error())
	}

	state, currentSharesNumber, minShares, totalShares := s.keyManager.Status()
	return &genkey.GetKeyStatusResult{
		IsLocked:      state == crypto.StateLocked,
		MinShares:     minShares,
		CurrentShares: currentSharesNumber,
		TotalShares:   totalShares,
	}, nil
}

func (s *KeyManagementService) AddShare(ctx context.Context, payload *genkey.AddSharePayload) (*genkey.AddShareResult, error) {
	state := s.keyManager.GetState()
	if state == crypto.StateUninitialized {
		return nil, genkey.NoKeySet("no master key configured")
	} else if state == crypto.StateUnlocked {
		return nil, genkey.KeyAlreadyUnlocked("key is already unlocked, cannot add share")
	}

	decodedShare, err := base64.StdEncoding.DecodeString(payload.Share)
	if err != nil {
		return nil, genkey.InternalError("error decoding share: " + err.Error())
	}

	index, unlocked, err := s.keyManager.AddShare(decodedShare)
	if err != nil {
		if errors.Is(err, crypto.ErrMaxSharesReached) {
			return nil, genkey.TooManyShares("maximum number of shares reached")
		}
		if errors.Is(err, crypto.ErrNoKeyConfigured) {
			return nil, genkey.NoKeySet("no master key configured")
		}
		if errors.Is(err, crypto.ErrCouldNotRecombine) {
			return nil, genkey.CouldNotRecombine("could not recombine shares: " + err.Error())
		}
		return nil, genkey.InternalError("error adding share: " + err.Error())
	}

	if unlocked {
		checksum, err := s.settingsRepository.GetSetting(ctx, columnMasterKeyChecksumColumn)
		if err != nil {
			return nil, genkey.InternalError("error retrieving master key checksum: " + err.Error())
		}

		decryptedChecksum, err := crypto.Decrypt(s.keyManager, checksum)
		if err != nil {
			s.keyManager.RollbackToLocked()
			return nil, genkey.WrongShares("error decrypting master key checksum: " + err.Error())
		}
		if string(decryptedChecksum) != checksumExpectedValue {
			s.keyManager.RollbackToLocked()
			return nil, genkey.WrongShares("master key checksum does not match expected value")
		}
	}

	return &genkey.AddShareResult{
		Index:    index,
		Unlocked: unlocked,
	}, nil
}

func (s *KeyManagementService) DeleteShare(_ context.Context, payload *genkey.DeleteSharePayload) error {
	state, _, _, maxShares := s.keyManager.Status()
	if state == crypto.StateUninitialized {
		return genkey.NoKeySet("no master key configured")
	} else if state == crypto.StateUnlocked {
		return genkey.KeyAlreadyUnlocked("key is already unlocked, cannot delete share")
	}

	if payload.Index < 0 || payload.Index >= maxShares {
		return genkey.WrongIndex("index provided does not match any share")
	}
	s.keyManager.RemoveShare(payload.Index)

	return nil
}
