package service

import (
	"context"
	"testing"

	"github.com/flussrd/fluss-back/accounts/models"
	rolesRepoMock "github.com/flussrd/fluss-back/accounts/repositories/roles/mock"
	usersRepoMock "github.com/flussrd/fluss-back/accounts/repositories/users/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	c := require.New(t)

	usersRepo := usersRepoMock.Repository{}
	rolesRepo := rolesRepoMock.Repository{}

	service := NewService(&usersRepo, &rolesRepo)

	user := models.User{
		UserID:      "USR1234",
		PhoneNumber: "+1809000000",
		Email:       "email@email.com",
	}

	ctx := context.Background()

	err := service.CreateUser(ctx, user)
	c.Nil(err)
}

func TestAddRoleToUser(t *testing.T) {
	c := require.New(t)

	usersRepo := usersRepoMock.Repository{}
	rolesRepo := rolesRepoMock.Repository{}

	service := NewService(&usersRepo, &rolesRepo)

	err := service.AddRoleToUser(context.Background(), "admin", "USR123")
	c.Nil(err)
}

func TestCreaterole(t *testing.T) {
	c := require.New(t)

	usersRepo := usersRepoMock.Repository{}
	rolesRepo := rolesRepoMock.Repository{}

	service := NewService(&usersRepo, &rolesRepo)

	err := service.CreateRole(context.Background(), models.Role{
		Name: "admin",
		Permissions: []models.Permission{
			{
				Resource: "modules",
				Action:   "*",
			},
		},
	})

	c.Nil(err)
}

func TestUpdateRole(t *testing.T) {
	c := require.New(t)

	usersRepo := usersRepoMock.Repository{}
	rolesRepo := rolesRepoMock.Repository{}

	service := NewService(&usersRepo, &rolesRepo)

	err := service.UpdateRole(context.Background(), models.Role{
		Name: "admin",
		Permissions: []models.Permission{
			{
				Resource: "modules",
				Action:   "*",
			},
		},
	})

	c.Nil(err)
}
