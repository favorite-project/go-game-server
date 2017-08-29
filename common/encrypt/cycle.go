package encrypt

import (
	"errors"
)

type Cycle struct {
	Key        []byte
	RandKeyLen int
}

func (cy *Cycle) Encrypt(data []byte) ([]byte, error) {
	if cy.Key == nil || data == nil {
		return nil, errors.New("Encrypt key or data nil")
	}

	if cy.RandKeyLen <= 0 {
		return nil, errors.New("Encrypt rand key length error")
	}

	cycle_key := RandKey(cy.RandKeyLen, KC_RAND_KIND_ALL)

	eDataLen := len(data) + cy.RandKeyLen
	eData := make([]byte, eDataLen)

	for i := 0; i < len(cycle_key); i++ {
		ri := i % len(cy.Key)
		eData[i] = cycle_key[i] ^ cy.Key[ri]
	}

	for i := cy.RandKeyLen; i < eDataLen; i++ {
		ri := i % cy.RandKeyLen
		eData[i] = cycle_key[ri] ^ data[i-cy.RandKeyLen]
	}

	return eData, nil
}

func (cy *Cycle) Decrypt(data []byte) ([]byte, error) {
	if cy.RandKeyLen <= 0 {
		return nil, errors.New("Decrypt rand key length error")
	}

	if cy.Key == nil || data == nil || len(data) <= cy.RandKeyLen {
		return nil, errors.New("Decrypt key or data error")
	}

	dData := make([]byte, len(data)-cy.RandKeyLen)
	cycle_key := make([]byte, cy.RandKeyLen)
	for i := 0; i < cy.RandKeyLen; i++ {
		ri := i % len(cy.Key)
		cycle_key[i] = data[i] ^ cy.Key[ri]
	}

	for i := cy.RandKeyLen; i < len(data); i++ {
		ri := i % cy.RandKeyLen
		dData[i-cy.RandKeyLen] = data[i] ^ cycle_key[ri]
	}

	return dData, nil
}
