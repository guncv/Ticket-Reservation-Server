package user

import (
	"errors"
	"unicode/utf8"

	"github.com/guncv/ticket-reservation-server/internal/domain/user/dto"
)

const (
	MinUserNameLength = 5
	MaxUserNameLength = 30
	MinPasswordLength = 8
	MaxPasswordLength = 30
)

func ValidateCreateUser(req dto.CreateUserReq, checkUserNameExists bool) error {
	if utf8.RuneCountInString(req.UserName) < MinUserNameLength {
		return errors.New("user name must be at least 5 characters")
	}

	if utf8.RuneCountInString(req.UserName) > MaxUserNameLength {
		return errors.New("user name must be at most 30 characters")
	}

	if utf8.RuneCountInString(req.Password) < MinPasswordLength {
		return errors.New("password must be at least 8 characters")
	}

	if utf8.RuneCountInString(req.Password) > MaxPasswordLength {
		return errors.New("password must be at most 30 characters")
	}

	if !checkUserNameExists {
		return errors.New("user name already exists")
	}

	return nil
}
