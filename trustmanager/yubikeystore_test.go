// +build pkcs11

package trustmanager

import (
	"crypto/rand"
	"testing"

	"github.com/docker/notary/passphrase"
	"github.com/docker/notary/tuf/data"
	"github.com/stretchr/testify/assert"
)

func clearAllKeys(t *testing.T) {
	// TODO(cyli): this is creating a new yubikey store because for some reason,
	// removing and then adding with the same YubiKeyStore causes
	// non-deterministic failures at least on Mac OS
	ret := passphrase.ConstantRetriever("passphrase")
	store, err := NewYubiKeyStore(NewKeyMemoryStore(ret), ret)
	assert.NoError(t, err)

	for k := range store.ListKeys() {
		err := store.RemoveKey(k)
		assert.NoError(t, err)
	}
}

func TestAddKeyToNextEmptyYubikeySlot(t *testing.T) {
	if !YubikeyAccessible() {
		t.Skip("Must have Yubikey access.")
	}
	clearAllKeys(t)

	ret := passphrase.ConstantRetriever("passphrase")
	store, err := NewYubiKeyStore(NewKeyMemoryStore(ret), ret)
	assert.NoError(t, err)
	SetYubikeyKeyMode(KeymodeNone)
	defer func() {
		SetYubikeyKeyMode(KeymodeTouch | KeymodePinOnce)
	}()

	keys := make([]string, 0, numSlots)

	// create the maximum number of keys
	for i := 0; i < numSlots; i++ {
		privKey, err := GenerateECDSAKey(rand.Reader)
		assert.NoError(t, err)

		err = store.AddKey(privKey.ID(), data.CanonicalRootRole, privKey)
		assert.NoError(t, err)

		keys = append(keys, privKey.ID())
	}

	listedKeys := store.ListKeys()
	assert.Len(t, listedKeys, numSlots)
	for _, k := range keys {
		r, ok := listedKeys[k]
		assert.True(t, ok)
		assert.Equal(t, data.CanonicalRootRole, r)
	}

	// numSlots is not actually the max - some keys might have more, so do not
	// test that adding more keys will fail.
}

// ImportKey imports a key as root without adding it to the backup store
func TestImportKey(t *testing.T) {
	if !YubikeyAccessible() {
		t.Skip("Must have Yubikey access.")
	}
	clearAllKeys(t)

	ret := passphrase.ConstantRetriever("passphrase")
	backup := NewKeyMemoryStore(ret)
	store, err := NewYubiKeyStore(backup, ret)
	assert.NoError(t, err)
	SetYubikeyKeyMode(KeymodeNone)
	defer func() {
		SetYubikeyKeyMode(KeymodeTouch | KeymodePinOnce)
	}()

	// generate key and import it
	privKey, err := GenerateECDSAKey(rand.Reader)
	assert.NoError(t, err)

	pemBytes, err := EncryptPrivateKey(privKey, "passphrase")
	assert.NoError(t, err)

	err = store.ImportKey(pemBytes, privKey.ID())
	assert.NoError(t, err)

	// key is not in backup store
	_, _, err = backup.GetKey(privKey.ID())
	assert.Error(t, err)

	// ensure key is in Yubikey -  create a new store, to make sure we're not
	// just using the keys cache
	store, err = NewYubiKeyStore(NewKeyMemoryStore(ret), ret)
	gottenKey, role, err := store.GetKey(privKey.ID())
	assert.NoError(t, err)
	assert.Equal(t, data.CanonicalRootRole, role)
	assert.Equal(t, privKey.Public(), gottenKey.Public())
}
