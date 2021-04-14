package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/flussrd/fluss-back/app/accounts/models"
	rolesRepository "github.com/flussrd/fluss-back/app/accounts/repositories/roles"
	usersRepository "github.com/flussrd/fluss-back/app/accounts/repositories/users"
	"github.com/flussrd/fluss-back/app/accounts/shared/httputils"
	"github.com/flussrd/fluss-back/app/accounts/shared/utils"
	"golang.org/x/crypto/bcrypt"
)

const (
	userIDPrefix = "USR"
)

var (
	// ErrMissingEmail missing email
	ErrMissingEmail = httputils.NewBadRequestError("missing email")
	// ErrMissingPassword missing password
	ErrMissingPassword = httputils.NewBadRequestError("missing password")
	// ErrMissingName missing name
	ErrMissingName = httputils.NewBadRequestError("missing name")
	// ErrMissingRole missing role
	ErrMissingRole = httputils.NewBadRequestError("missing role")
	// ErrPasswordHashingFailed password hashing failed
	ErrPasswordHashingFailed = errors.New("hashing password failed")
	// ErrGeneratingIDFailed generating id failed
	ErrGeneratingIDFailed = errors.New("generating id failed")
	// ErrValidatingRoleFailed validating role failed
	ErrValidatingRoleFailed = errors.New("validating role failed")
	// ErrRoleNotValid role not valid
	ErrRoleNotValid = httputils.NewBadRequestError("role not valid")
	// ErrMissingRoleName missing role name
	ErrMissingRoleName = httputils.NewBadRequestError("missing role name")
	// ErrMissingPermissions missing permissions
	ErrMissingPermissions = httputils.NewBadRequestError("missing permissions")
	// ErrMissingActionInPermission missing action in permissions
	ErrMissingActionInPermission = httputils.NewBadRequestError("missing actions in permission")
	// ErrMissingResourceInPermission missing resource in permission
	ErrMissingResourceInPermission = httputils.NewBadRequestError("missing resource in permission")
	// ErrInvalidCredentials invalid credentials
	ErrInvalidCredentials = httputils.ErrorResponse{Code: http.StatusUnauthorized, Message: "invalid credentials"}
)

var (
	generatePasswordHashFunction func(password []byte, cost int) ([]byte, error)
	generateIDFunction           func(prefix string) (string, error)
)

func init() {
	generatePasswordHashFunction = bcrypt.GenerateFromPassword
	generateIDFunction = utils.GenerateID
}

type service struct {
	usersRepo usersRepository.Repository
	rolesRepo rolesRepository.Repository
}

// NewService returns a new service entity to be able to execuse business logic
func NewService(usersRepo usersRepository.Repository, rolesRepo rolesRepository.Repository) Service {
	return service{
		usersRepo: usersRepo,
		rolesRepo: rolesRepo,
	}
}

// CreateUser creates a new user
func (s service) CreateUser(ctx context.Context, user models.User) (models.User, error) {
	err := validateCreateUserParams(user)
	if err != nil {
		return models.User{}, err
	}

	isValid, err := s.isValidRole(ctx, user.RoleName)
	if err != nil {
		return models.User{}, err
	}

	if !isValid {
		return models.User{}, ErrRoleNotValid
	}

	hashedPassword, err := generatePasswordHashFunction([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, ErrPasswordHashingFailed
	}

	user.Password = string(hashedPassword)

	id, err := generateIDFunction(userIDPrefix)
	if err != nil {
		return models.User{}, ErrGeneratingIDFailed
	}

	user.UserID = id

	insertedUser, err := s.usersRepo.SaveUser(ctx, user)
	if errors.As(err, &usersRepository.ErrDuplicateFields{}) {
		return models.User{}, httputils.NewBadRequestError(err.Error())
	}

	return insertedUser, nil
}

func (s service) isValidRole(ctx context.Context, roleName string) (bool, error) {
	_, err := s.rolesRepo.GetRole(ctx, roleName)
	if errors.Is(err, rolesRepository.ErrNotFound) {
		return false, nil
	}

	if err != nil {
		return false, ErrValidatingRoleFailed
	}

	return true, nil
}

func validateCreateUserParams(user models.User) error {
	if user.Email == "" {
		return ErrMissingEmail
	}

	if user.Password == "" {
		return ErrMissingPassword
	}

	if user.Name == "" {
		return ErrMissingName
	}

	if user.RoleName == "" {
		return ErrMissingRole
	}

	return nil
}

// AddRoleToUser adds a role to a user
func (s service) AddRoleToUser(ctx context.Context, roleName string, userID string) error {
	return nil
}

// CreateRole creates a new role
func (s service) CreateRole(ctx context.Context, role models.Role) error {
	err := validateCreateRoleParams(role)
	if err != nil {
		return err
	}

	return s.rolesRepo.CreateRole(ctx, role)
}

func validateCreateRoleParams(role models.Role) error {
	if role.Name == "" {
		return ErrMissingRoleName
	}

	if len(role.Permissions) == 0 {
		return ErrMissingPermissions
	}

	for _, permission := range role.Permissions {
		if permission.Action == "" {
			return ErrMissingActionInPermission
		}

		if permission.Resource == "" {
			return ErrMissingResourceInPermission
		}
	}

	return nil
}

// UpdateRole updates a role
func (s service) UpdateRole(ctx context.Context, role models.Role) error {
	return nil
}

// GetRoles returns all the roles
func (s service) GetRoles(ctx context.Context) ([]models.Role, error) {
	// TODO: handle repo specific errors
	return s.rolesRepo.GetRoles(ctx)
}

// Login authenticates a user
func (s service) Login(ctx context.Context, email string, password string) (LoginResponse, error) {
	err := validateLoginInput(email, password)
	if err != nil {
		return LoginResponse{}, err
	}

	user, err := s.usersRepo.GetUserByEmail(ctx, email)
	if errors.Is(err, usersRepository.ErrNotFound) {
		return LoginResponse{}, ErrInvalidCredentials
	}

	if !isPasswordCorrect(password, user.Password) {
		return LoginResponse{}, ErrInvalidCredentials
	}

	token, err := generateToken(user)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{Token: token}, nil
}

func generateToken(user models.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"roleName": user.RoleName,
		"sub":      user.UserID,
		"iss":      "fluss-back", // TODO: make this a const
		"iat":      time.Now(),
		"exp":      time.Now().Add(time.Hour * 24), // TODO: make the adding a const
	})

	signedToken, err := token.SignedString([]byte(os.Getenv("TOKEN_SECRET")))
	if err != nil {
		return "", fmt.Errorf("error signing token: " + err.Error())
	}

	return signedToken, nil
}

func isPasswordCorrect(enteredPassword string, storedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(enteredPassword))

	return err == nil
}

func validateLoginInput(email, password string) error {
	if email == "" {
		return ErrMissingEmail
	}

	if password == "" {
		return ErrMissingPassword
	}

	return nil
}
