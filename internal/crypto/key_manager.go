package crypto

import (
	"errors"
	"fmt"
	"sync"
)

var (
	ErrKeyLocked         = errors.New("master key is locked")
	ErrNoKeyConfigured   = errors.New("no key system configured")
	ErrInvalidShares     = errors.New("invalid number of shares")
	ErrMaxSharesReached  = errors.New("maximum number of shares reached")
	ErrCouldNotRecombine = errors.New("could not recombine shares to unlock key")
)

type State int

const (
	StateUninitialized State = iota
	StateLocked
	StateUnlocked
)

type KeyManager struct {
	mu        sync.RWMutex
	state     State
	masterKey []byte
	shares    [][]byte
	minShares int
	maxShares int
}

var (
	instance *KeyManager
	once     sync.Once
)

// GetKeyManager returns the singleton instance of the key manager
func GetKeyManager() *KeyManager {
	once.Do(func() {
		instance = &KeyManager{
			state: StateUninitialized,
		}
	})
	return instance
}

// ConfigureKeySystem initializes the system with share thresholds, called only if a key was previously configured
func (km *KeyManager) ConfigureKeySystem(minShares, maxShares int) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	if minShares <= 0 || maxShares < minShares {
		return ErrInvalidShares
	}

	km.minShares = minShares
	km.maxShares = maxShares
	km.shares = [][]byte{}
	km.masterKey = nil
	km.state = StateLocked

	return nil
}

func (km *KeyManager) SetNewMasterKey(masterKey []byte, minShares, maxShares int) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	if minShares <= 0 || maxShares < minShares {
		return ErrInvalidShares
	}

	km.masterKey = make([]byte, len(masterKey))
	copy(km.masterKey, masterKey)
	km.minShares = minShares
	km.maxShares = maxShares
	km.shares = [][]byte{}
	km.state = StateUnlocked

	return nil
}

// AddShare adds a share and attempts to unlock the key if enough are present
// Returns the index of the added share
func (km *KeyManager) AddShare(share []byte) (index int, unlocked bool, err error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	if km.state == StateUninitialized {
		return -1, false, ErrNoKeyConfigured
	}
	if len(km.shares) >= km.maxShares {
		return -1, false, ErrMaxSharesReached
	}

	km.shares = append(km.shares, share)

	if len(km.shares) >= km.minShares && km.state != StateUnlocked {
		masterKey, err := CombineShares(km.shares)
		if err != nil {
			km.shares = km.shares[:len(km.shares)-1]
			return -1, false, fmt.Errorf("%w: %v", ErrCouldNotRecombine, err)
		}
		if len(masterKey) == 0 {
			km.shares = km.shares[:len(km.shares)-1]
			return -1, false, fmt.Errorf("%w: no valid master key reconstructed", ErrCouldNotRecombine)
		}
		km.masterKey = make([]byte, len(masterKey))
		copy(km.masterKey, masterKey)
		km.state = StateUnlocked
		return len(km.shares) - 1, true, nil
	}

	return len(km.shares) - 1, false, nil
}

// RemoveShare removes a share at the given index.
func (km *KeyManager) RemoveShare(index int) {
	km.mu.Lock()
	defer km.mu.Unlock()

	if index < 0 || index >= len(km.shares) {
		return
	}

	km.shares = append(km.shares[:index], km.shares[index+1:]...)
}

// GetMasterKey returns the unlocked master key
func (km *KeyManager) GetMasterKey() ([]byte, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.state != StateUnlocked {
		return nil, ErrKeyLocked
	}

	keyCopy := make([]byte, len(km.masterKey))
	copy(keyCopy, km.masterKey)
	return keyCopy, nil
}

func (km *KeyManager) GetState() State {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.state
}

// RollbackToLocked rollback state to locked, called when key is wrong
func (km *KeyManager) RollbackToLocked() {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.state = StateLocked
	km.masterKey = nil
}

func (km *KeyManager) Status() (state State, currentSharesNumber int, minShares int, maxShares int) {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.state, len(km.shares), km.minShares, km.maxShares
}
