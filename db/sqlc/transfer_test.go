package db

import (
	"context"
	"github.com/SaishNaik/simplebank/utils"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func CreateRandomTransfer(t *testing.T, account1, account2 Account) Transfer {
	ctx := context.Background()
	amount := utils.RandomMoney()
	arg := CreateTransferParams{
		FromAccountID: account1.ID,
		ToAccountID:   account2.ID,
		Amount:        amount,
	}
	transfer, err := testStore.CreateTransfer(ctx, arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)
	return transfer
}

func TestCreateTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	CreateRandomTransfer(t, account1, account2)
}

func TestGetTransfer(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	expectedTransfer := CreateRandomTransfer(t, account1, account2)
	gotTransfer, err := testStore.GetTransfer(context.Background(), expectedTransfer.ID)
	require.NoError(t, err)
	require.NotEmpty(t, gotTransfer)

	require.Equal(t, expectedTransfer.FromAccountID, gotTransfer.FromAccountID)
	require.Equal(t, expectedTransfer.ToAccountID, gotTransfer.ToAccountID)
	require.Equal(t, expectedTransfer.Amount, gotTransfer.Amount)
	require.Equal(t, expectedTransfer.ID, gotTransfer.ID)
	require.WithinDuration(t, expectedTransfer.CreatedAt, gotTransfer.CreatedAt, time.Second)
}

func TestListTransfers(t *testing.T) {
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	for i := 0; i < 5; i++ {
		CreateRandomTransfer(t, account1, account2)
		CreateRandomTransfer(t, account2, account1)
	}
	transfers, err := testStore.ListTransfers(context.Background(), ListTransfersParams{
		FromAccountID: account1.ID,
		ToAccountID:   account1.ID,
		Limit:         5,
		Offset:        5,
	})

	require.NoError(t, err)
	require.Len(t, transfers, 5)
	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
		require.True(t, transfer.FromAccountID == account1.ID || transfer.ToAccountID == account1.ID)
	}
}
