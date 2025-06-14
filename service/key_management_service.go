package service

import (
	"context"
	"encoding/base64"
	"errors"
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
	columnTotalShares       = "total_shares"
	columnMinShares         = "min_shares"
	columnMasterKeyChecksum = "master_key_checksum"
	checksumExpectedValue   = "fishykeys_checksum"
)

type KeyManagementService struct {
	keyManager         *crypto.KeyManager
	settingsRepository repository.GlobalSettingsRepository
}

func NewKeyManagementService(keyManager *crypto.KeyManager, repo repository.GlobalSettingsRepository) *KeyManagementService {
	keySettings, err := repo.GetSettings(context.Background(), columnMasterKeyChecksum, columnTotalShares, columnMinShares)
	if err != nil {
		if !errors.Is(err, repository.ErrSettingNotFound) {
			log.Fatalf("error retrieving key settings on service init: %v", err)
		}
	} else {
		minShares, err := strconv.Atoi(keySettings[columnMinShares])
		if err != nil {
			log.Fatalf("error parsing min shares from db: %v", err)
		}
		totalShares, err := strconv.Atoi(keySettings[columnTotalShares])
		if err != nil {
			log.Fatalf("error parsing total shares from db: %v", err)
		}
		err = keyManager.ConfigureKeySystem(minShares, totalShares)
		if err != nil {
			log.Fatalf("error configuring key system with existing shares: %v", err)
		}
	}

	return &KeyManagementService{
		keyManager:         keyManager,
		settingsRepository: repo,
	}
}

func (s *KeyManagementService) CreateMasterKey(ctx context.Context, payload *genkey.CreateMasterKeyPayload) (*genkey.CreateMasterKeyResult, error) {
	if payload.TotalShares <= 0 || payload.MinShares <= 0 || payload.MinShares > payload.TotalShares {
		return nil, ErrInvalidParameters
	}

	_, err := s.settingsRepository.GetSetting(ctx, columnMasterKeyChecksum)
	if err == nil {
		return nil, genkey.KeyAlreadyExists("master key already exists")
	}

	masterKey, err := crypto.GenerateSecret()
	if err != nil {
		return nil, genkey.InternalError("error generating master key: " + err.Error())
	}
	checksum, err := crypto.EncryptWithKey(masterKey, []byte(checksumExpectedValue))
	if err != nil {
		return nil, genkey.InternalError("error encrypting master key checksum: " + err.Error())
	}
	shares, err := crypto.SplitSecret(masterKey, payload.TotalShares, payload.MinShares)
	if err != nil {
		return nil, genkey.InternalError("error splitting secret into shares: " + err.Error())
	}

	encodedShares := make([]string, len(shares))
	for i, b := range shares {
		encodedShares[i] = base64.StdEncoding.EncodeToString(b)
	}

	err = s.settingsRepository.StoreSettings(ctx, map[string]string{
		columnTotalShares:       strconv.Itoa(payload.TotalShares),
		columnMinShares:         strconv.Itoa(payload.MinShares),
		columnMasterKeyChecksum: checksum,
	})
	if err != nil {
		return nil, genkey.InternalError("error storing key settings: " + err.Error())
	}

	err = s.keyManager.SetNewMasterKey(masterKey, payload.MinShares, payload.TotalShares)
	if err != nil {
		return nil, genkey.InternalError("error setting new master key: " + err.Error())
	}

	return &genkey.CreateMasterKeyResult{
		Shares: encodedShares,
	}, nil
}

func (s *KeyManagementService) GetKeyStatus(ctx context.Context) (*genkey.GetKeyStatusResult, error) {
	_, err := s.settingsRepository.GetSetting(ctx, columnMasterKeyChecksum)
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
		checksum, err := s.settingsRepository.GetSetting(ctx, columnMasterKeyChecksum)
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
