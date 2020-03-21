package contracts

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/dapperlabs/flow-go-sdk"
	"github.com/dapperlabs/flow-go-sdk/keys"
	"github.com/dapperlabs/flow-go-sdk/utils/examples"
)

const (
	fungibleTokenTutorialContractFile = "./contracts/fungible-token-tutorial.cdc"
)

func TestFungibleTokenTutorialContractDeployment(t *testing.T) {
	b := examples.NewEmulator()

	// Should be able to deploy a contract as a new account with no keys.
	tokenCode := examples.ReadFile(fungibleTokenTutorialContractFile)
	_, err := b.CreateAccount(nil, tokenCode, examples.GetNonce())
	assert.NoError(t, err)

	_, err = b.CommitBlock()
	assert.NoError(t, err)
}

func TestFungibleTokenTutorialContractCreation(t *testing.T) {
	b := examples.NewEmulator()

	// First, *update* the contract
	tokenCode := examples.ReadFile(fungibleTokenTutorialContractFile)
	err := b.UpdateAccountCode(tokenCode, examples.GetNonce())
	assert.NoError(t, err)

	t.Run("Set up account 1", func(t *testing.T) {
		tx := flow.Transaction{
			Script: []byte(
				fmt.Sprintf(
					`
                      import FungibleToken from 0x%s

                      transaction {
                          prepare(acct: AuthAccount) {
                              acct.published[&AnyResource{FungibleToken.Receiver}] =
                                   &acct.storage[FungibleToken.Vault] as &AnyResource{FungibleToken.Receiver}

                              acct.storage[&FungibleToken.Vault] =
                                   &acct.storage[FungibleToken.Vault] as &FungibleToken.Vault
                          }
                      }
	               `,
					b.RootAccountAddress().Short(),
				),
			),
			Nonce:          examples.GetNonce(),
			ComputeLimit:   10,
			PayerAccount:   b.RootAccountAddress(),
			ScriptAccounts: []flow.Address{b.RootAccountAddress()},
		}

		examples.SignAndSubmit(t, b, tx, []flow.AccountPrivateKey{b.RootKey()}, []flow.Address{b.RootAccountAddress()}, false)
	})

	var account2Address flow.Address

	t.Run("Create account 2", func(t *testing.T) {

		var err error
		publicKeys := []flow.AccountPublicKey{b.RootKey().PublicKey(keys.PublicKeyWeightThreshold)}
		account2Address, err = b.CreateAccount(publicKeys, nil, examples.GetNonce())
		assert.NoError(t, err)
	})

	t.Run("Set up account 2", func(t *testing.T) {
		tx := flow.Transaction{
			Script: []byte(
				fmt.Sprintf(
					`
                      // NOTE: using different import address to ensure user can use different formats
                      import FungibleToken from 0x00%s

                      transaction {

                          prepare(acct: AuthAccount) {
                              // create a new vault instance
                              let vaultA <- FungibleToken.createEmptyVault()

                              // store it in the account storage
                              // and destroy whatever was there previously
                              let oldVault <- acct.storage[FungibleToken.Vault] <- vaultA
                              destroy oldVault

                              acct.published[&AnyResource{FungibleToken.Receiver}] =
                                  &acct.storage[FungibleToken.Vault] as &AnyResource{FungibleToken.Receiver}

                              acct.storage[&FungibleToken.Vault] =
                                  &acct.storage[FungibleToken.Vault] as &FungibleToken.Vault
                          }
                      }
                    `,
					b.RootAccountAddress().Short(),
				),
			),
			Nonce:          examples.GetNonce(),
			ComputeLimit:   10,
			PayerAccount:   account2Address,
			ScriptAccounts: []flow.Address{account2Address},
		}

		examples.SignAndSubmit(t, b, tx, []flow.AccountPrivateKey{b.RootKey()}, []flow.Address{account2Address}, false)
	})
}