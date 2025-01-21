package db

import (
	"context"
	"github.com/SaishNaik/simplebank/utils"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomUser(t *testing.T) User {

	hashedPassword, err := utils.HashPassword(utils.RandomString(6))
	require.NoError(t, err)

	params := CreateUserParams{
		Username:       utils.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       utils.RandomOwner(),
		Email:          utils.RandomEmail(),
	}
	ctx := context.Background()
	user, err := testStore.CreateUser(ctx, params)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, params.Username, user.Username)
	require.Equal(t, params.HashedPassword, user.HashedPassword)
	require.Equal(t, params.Username, user.Username)
	require.Equal(t, params.Email, user.Email)
	require.True(t, user.PasswordChangedAt.IsZero())
	require.NotZero(t, user.CreatedAt)
	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	ctx := context.Background()
	createdUser := createRandomUser(t)
	gotUser, err := testStore.GetUser(ctx, createdUser.Username)
	require.NoError(t, err)
	require.NotEmpty(t, gotUser)
	require.Equal(t, createdUser.Username, gotUser.Username)
	require.Equal(t, createdUser.HashedPassword, gotUser.HashedPassword)
	require.Equal(t, createdUser.Username, gotUser.Username)
	require.Equal(t, createdUser.Email, gotUser.Email)
	require.WithinDuration(t, createdUser.CreatedAt, gotUser.CreatedAt, time.Second)
	require.WithinDuration(t, createdUser.PasswordChangedAt, gotUser.PasswordChangedAt, time.Second)
}

func TestUpdateUserOnlyFullName(t *testing.T) {
	oldUser := createRandomUser(t)
	newFullName := utils.RandomOwner()
	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		FullName: pgtype.Text{
			Valid:  true,
			String: newFullName,
		},
		Username: oldUser.Username,
	})
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.NotEqual(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}

func TestUpdateUserOnlyEmail(t *testing.T) {
	oldUser := createRandomUser(t)
	newEmail := utils.RandomEmail()
	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		Email: pgtype.Text{
			Valid:  true,
			String: newEmail,
		},
		Username: oldUser.Username,
	})
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, newEmail, updatedUser.Email)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
}

func TestUpdateUserOnlyHashedPassword(t *testing.T) {
	oldUser := createRandomUser(t)
	newHashedPassword, err := utils.HashPassword(utils.RandomString(7))
	require.NoError(t, err)
	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		HashedPassword: pgtype.Text{
			Valid:  true,
			String: newHashedPassword,
		},
		Username: oldUser.Username,
	})
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)
}

func TestUpdateUserAllFields(t *testing.T) {
	oldUser := createRandomUser(t)
	newFullName := utils.RandomOwner()
	newEmail := utils.RandomEmail()
	newHashedPassword, err := utils.HashPassword(utils.RandomString(7))
	require.NoError(t, err)
	updatedUser, err := testStore.UpdateUser(context.Background(), UpdateUserParams{
		HashedPassword: pgtype.Text{
			Valid:  true,
			String: newHashedPassword,
		},
		FullName: pgtype.Text{
			Valid:  true,
			String: newFullName,
		},
		Email: pgtype.Text{
			Valid:  true,
			String: newEmail,
		},
		Username: oldUser.Username,
	})
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.NotEqual(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.NotEqual(t, oldUser.Email, updatedUser.Email)
	require.NotEqual(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, newHashedPassword, updatedUser.HashedPassword)
	require.Equal(t, newFullName, updatedUser.FullName)
	require.Equal(t, newEmail, updatedUser.Email)
}
