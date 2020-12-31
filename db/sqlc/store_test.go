package db

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)

	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	n := 5
	amount := int64(10)
	errChan := make(chan error)
	resultChan := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		txName := fmt.Sprintf("tx %d", i+1)
		go func() {
			ctx := context.WithValue(context.Background(), txKey, txName)
			result, err := store.TransferTx(ctx, TransferTxParams {
				FromAccountID: account1.ID,
				ToAccountID: account2.ID,
				Amount: amount,
			})

			errChan <- err
			resultChan <- result
		}()
	}

	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <- errChan
		require.NoError(t, err)

		result := <- resultChan
		require.NotEmpty(t, result)

		Transfer := result.Transfer
		require.NotEmpty(t, Transfer)
		require.Equal(t, account1.ID, Transfer.FromAccountID)
		require.Equal(t, account2.ID, Transfer.ToAccountID)
		require.Equal(t, amount, Transfer.Amount)
		require.NotZero(t, Transfer.ID)
		require.NotZero(t, Transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), Transfer.ID)
		require.NoError(t, err)

		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		// check account
		fromAccount := result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1 % amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)
	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance - amount * int64(n), updatedAccount1.Balance)
	require.Equal(t, account2.Balance + amount * int64(n), updatedAccount2.Balance)
}
