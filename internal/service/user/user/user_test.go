package user

import (
	"strings"
	"testing"

	"github.com/guncv/ticket-reservation-server/internal/service/user/dto"
	"github.com/stretchr/testify/require"
)

func TestValidateCreateUser(t *testing.T) {
	testCases := []struct {
		name                string
		req                 func() dto.CreateUserReq
		checkUserNameExists bool
		verify              func(t *testing.T, actual error)
	}{
		{
			name: "Success",
			req: func() dto.CreateUserReq {
				return dto.CreateUserReq{
					UserName: "test_user",
					Password: "test_password",
				}
			},
			checkUserNameExists: false,
			verify: func(t *testing.T, actual error) {
				require.NoError(t, actual)
			},
		},
		{
			name: "Success_EdgeCase_MinLength",
			req: func() dto.CreateUserReq {
				return dto.CreateUserReq{
					UserName: "abcde",
					Password: "12345678",
				}
			},
			checkUserNameExists: false,
			verify: func(t *testing.T, actual error) {
				require.NoError(t, actual)
			},
		},
		{
			name: "Success_EdgeCase_MaxLength",
			req: func() dto.CreateUserReq {
				return dto.CreateUserReq{
					UserName: strings.Repeat("a", 30),
					Password: strings.Repeat("b", 30),
				}
			},
			checkUserNameExists: false,
			verify: func(t *testing.T, actual error) {
				require.NoError(t, actual)
			},
		},
		{
			name: "Error_UserNameTooShort",
			req: func() dto.CreateUserReq {
				return dto.CreateUserReq{
					UserName: "ab",
					Password: "validpass",
				}
			},
			checkUserNameExists: false,
			verify: func(t *testing.T, actual error) {
				require.Error(t, actual)
				require.Contains(t, actual.Error(), "user name must be at least 5 characters")
			},
		},
		{
			name: "Error_UserNameTooLong",
			req: func() dto.CreateUserReq {
				return dto.CreateUserReq{
					UserName: strings.Repeat("a", 31),
					Password: "validpassword",
				}
			},
			checkUserNameExists: false,
			verify: func(t *testing.T, actual error) {
				require.Error(t, actual)
				require.Contains(t, actual.Error(), "user name must be at most 30 characters")
			},
		},
		{
			name: "Error_PasswordTooShort",
			req: func() dto.CreateUserReq {
				return dto.CreateUserReq{
					UserName: "valid_user",
					Password: "short",
				}
			},
			checkUserNameExists: false,
			verify: func(t *testing.T, actual error) {
				require.Error(t, actual)
				require.Contains(t, actual.Error(), "password must be at least 8 characters")
			},
		},
		{
			name: "Error_PasswordTooLong",
			req: func() dto.CreateUserReq {
				return dto.CreateUserReq{
					UserName: "valid_user",
					Password: strings.Repeat("x", 31),
				}
			},
			checkUserNameExists: false,
			verify: func(t *testing.T, actual error) {
				require.Error(t, actual)
				require.Contains(t, actual.Error(), "password must be at most 30 characters")
			},
		},
		{
			name: "Error_UserNameAlreadyExists",
			req: func() dto.CreateUserReq {
				return dto.CreateUserReq{
					UserName: "existing_user",
					Password: "validpassword",
				}
			},
			checkUserNameExists: true,
			verify: func(t *testing.T, actual error) {
				require.Error(t, actual)
				require.Contains(t, actual.Error(), "user name already exists")
			},
		},
		{
			name: "Error_UserNameEmpty",
			req: func() dto.CreateUserReq {
				return dto.CreateUserReq{
					UserName: "",
					Password: "validpassword",
				}
			},
			checkUserNameExists: false,
			verify: func(t *testing.T, actual error) {
				require.Error(t, actual)
				require.Contains(t, actual.Error(), "user name must be at least 5 characters")
			},
		},
		{
			name: "Error_UserNameTooShort_Unicode",
			req: func() dto.CreateUserReq {
				return dto.CreateUserReq{
					UserName: "用户",
					Password: "validpassword",
				}
			},
			checkUserNameExists: false,
			verify: func(t *testing.T, actual error) {
				require.Error(t, actual)
				require.Contains(t, actual.Error(), "user name must be at least 5 characters")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := ValidateCreateUser(tc.req(), tc.checkUserNameExists)
			tc.verify(t, actual)
		})
	}
}
