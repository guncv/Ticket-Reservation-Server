package user

import (
	"errors"
	"unicode/utf8"

	"github.com/guncv/ticket-reservation-server/internal/service/user/dto"
)

const (
	MinUserNameLength = 5
	MaxUserNameLength = 30
	MinPasswordLength = 8
	MaxPasswordLength = 30
)

func validateCredentials(userName, password string) error {
	if utf8.RuneCountInString(userName) < MinUserNameLength {
		return errors.New("user name must be at least 5 characters")
	}

	if utf8.RuneCountInString(userName) > MaxUserNameLength {
		return errors.New("user name must be at most 30 characters")
	}

	if utf8.RuneCountInString(password) < MinPasswordLength {
		return errors.New("password must be at least 8 characters")
	}

	if utf8.RuneCountInString(password) > MaxPasswordLength {
		return errors.New("password must be at most 30 characters")
	}

	return nil
}

func ValidateCreateUser(req dto.CreateUserReq, checkUserNameExists bool) error {
	if err := validateCredentials(req.UserName, req.Password); err != nil {
		return err
	}

	if checkUserNameExists {
		return errors.New("user name already exists")
	}

	return nil
}

func ValidateLoginUser(req dto.LoginUserReq) error {
	return validateCredentials(req.UserName, req.Password)
}
