package service

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/Vidalee/FishyKeys/backend/repository"
	"log"
	"strconv"

	genkey "github.com/Vidalee/FishyKeys/backend/gen/fishykeys"
	"github.com/Vidalee/FishyKeys/backend/internal/crypto"
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
		return nil, err
	}
	checksum, err := crypto.Encrypt(masterKey, []byte(checksumExpectedValue))
	shares, err := crypto.SplitSecret(masterKey, payload.TotalShares, payload.MinShares)
	if err != nil {
		return nil, err
	}

	encodedShares := make([]string, len(shares))
	for i, b := range shares {
		encodedShares[i] = base64.StdEncoding.EncodeToString(b)
	}

	err = s.settingsRepository.StoreSettings(ctx, map[string]string{
		columnTotalShares:       strconv.Itoa(payload.TotalShares),
		columnMinShares:         strconv.Itoa(payload.MinShares),
		columnMasterKeyChecksum: base64.StdEncoding.EncodeToString(checksum),
	})
	if err != nil {
		return nil, err
	}

	err = s.keyManager.SetNewMasterKey(masterKey, payload.MinShares, payload.TotalShares)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	state, currentSharesNumber, minShares, totalShares := s.keyManager.Status()
	return &genkey.GetKeyStatusResult{
		IsLocked:      state == crypto.StateLocked,
		MinShares:     minShares,
		CurrentShares: currentSharesNumber,
		TotalShares:   totalShares,
	}, nil
}

func (s *KeyManagementService) AddShare() {
	//s.keyManager.AddShare()
}

func (s *KeyManagementService) RemoveShare() {
	//s.keyManager.RemoveShare()
}
