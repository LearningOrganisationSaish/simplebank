package db

import (
	"context"
	"github.com/SaishNaik/simplebank/utils"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomAccount(t *testing.T) Account {
	user := createRandomUser(t)
	params := CreateAccountParams{
		Owner:    user.Username,
		Balance:  utils.RandomMoney(),
		Currency: utils.RandomCurrency(),
	}
	ctx := context.Background()
	account, err := testStore.CreateAccount(ctx, params)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, params.Owner, account.Owner)
	require.Equal(t, params.Balance, account.Balance)
	require.Equal(t, params.Currency, account.Currency)
	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)
	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	ctx := context.Background()
	createdAccount := createRandomAccount(t)
	gotAccount, err := testStore.GetAccount(ctx, createdAccount.ID)
	require.NoError(t, err)
	require.NotEmpty(t, gotAccount)
	require.Equal(t, createdAccount.Owner, gotAccount.Owner)
	require.Equal(t, createdAccount.Balance, gotAccount.Balance)
	require.Equal(t, createdAccount.Currency, gotAccount.Currency)
	require.Equal(t, createdAccount.ID, gotAccount.ID)
	require.WithinDuration(t, createdAccount.CreatedAt, gotAccount.CreatedAt, time.Second)
}

func TestUpdateAccount(t *testing.T) {
	createdAccount := createRandomAccount(t)
	expectedBalance := utils.RandomMoney()
	params := UpdateAccountParams{
		ID:      createdAccount.ID,
		Balance: expectedBalance,
	}
	gotAccount, err := testStore.UpdateAccount(context.Background(), params)
	require.NoError(t, err)
	require.NotEmpty(t, gotAccount)

	require.Equal(t, createdAccount.Owner, gotAccount.Owner)
	require.Equal(t, createdAccount.Currency, gotAccount.Currency)
	require.Equal(t, createdAccount.ID, gotAccount.ID)
	require.WithinDuration(t, createdAccount.CreatedAt, gotAccount.CreatedAt, time.Second)
	require.Equal(t, expectedBalance, gotAccount.Balance)
}

func TestDeleteAccount(t *testing.T) {
	createdAccount := createRandomAccount(t)
	err := testStore.DeleteAccount(context.Background(), createdAccount.ID)
	require.NoError(t, err)
	gotAccount, err := testStore.GetAccount(context.Background(), createdAccount.ID)
	require.Error(t, err)
	require.Equal(t, err, ErrRecordNotFound)
	require.Empty(t, gotAccount)
}

func TestListAccount(t *testing.T) {
	var lastAccount Account
	for i := 0; i < 10; i++ {
		lastAccount = createRandomAccount(t)
	}
	params := ListAccountsParams{
		Owner:  lastAccount.Owner,
		Limit:  5,
		Offset: 0,
	}
	accounts, err := testStore.ListAccounts(context.Background(), params)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)
	for i := 0; i < len(accounts); i++ {
		require.NotEmpty(t, accounts[i])
		require.Equal(t, lastAccount.Owner, accounts[i].Owner)
	}
}
