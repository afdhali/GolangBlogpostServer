package unittest

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/afdhali/GolangBlogpostServer/internal/entity"
	"github.com/afdhali/GolangBlogpostServer/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestUserRepository_Create(t *testing.T) {
	// Create sqlmock
	dbMock, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	defer dbMock.Close()

	// Create GORM DB with mock
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: dbMock,
	}), &gorm.Config{})
	require.NoError(t, err)

	// Create repository
	repo := repository.NewUserRepository(gormDB)

	ctx := context.Background()
	user := &entity.User{
		BaseEntity: entity.BaseEntity{ID: uuid.Nil},
		Username:   "testuser",
		Email:      "test@example.com",
		Password:   "hashedpass",
		FullName:   "",
		Role:       entity.RoleUser,
		IsActive:   true,
		Avatar:     "",
	}

	// Mock the transaction Begin
	sqlMock.ExpectBegin()

	// Mock the INSERT query with the actual order from log, including "avatar"
	sqlMock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO "users" ("created_at","updated_at","deleted_at","username","email","password","full_name","role","is_active","avatar","id") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING "id"`)).
		WithArgs(
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
			nil,              // deleted_at
			user.Username,
			user.Email,
			user.Password,
			user.FullName,
			string(user.Role),
			user.IsActive,
			user.Avatar,      // avatar
			sqlmock.AnyArg(), // id
		).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))

	// Mock Commit
	sqlMock.ExpectCommit()

	// Call Create
	err = repo.Create(ctx, user)
	require.NoError(t, err)

	// Verify ID was set by BeforeCreate
	require.NotEqual(t, uuid.Nil, user.ID)

	// Verify expectations
	require.NoError(t, sqlMock.ExpectationsWereMet())
}