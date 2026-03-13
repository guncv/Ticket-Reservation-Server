package user

import (
	"context"

	"github.com/guncv/ticket-reservation-server/internal/domain/user/dto"
	"github.com/guncv/ticket-reservation-server/internal/domain/user/repo"
	"github.com/guncv/ticket-reservation-server/internal/domain/user/user"
)

func (s *userService) CreateUser(ctx context.Context, req dto.CreateUserReq) (dto.CreateUserResp, error) {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return dto.CreateUserResp{}, err
	}
	defer tx.Rollback(ctx)

	exists, err := s.userRepo.CheckUserNameExists(ctx, req.UserName)
	if err != nil {
		return dto.CreateUserResp{}, err
	}

	err = user.ValidateCreateUser(req, exists)
	if err != nil {
		return dto.CreateUserResp{}, err
	}

	hashedPassword, err := user.HashPassword(req.Password)
	if err != nil {
		return dto.CreateUserResp{}, err
	}

	userID, err := s.userRepo.CreateUser(ctx, repo.CreateUserParams{
		UserName:       req.UserName,
		HashedPassword: hashedPassword,
		Role:           dto.UserRoleUser,
	})
	if err != nil {
		return dto.CreateUserResp{}, err
	}

	session, err := s.CreateSession(ctx, userID)
	if err != nil {
		return dto.CreateUserResp{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return dto.CreateUserResp{}, err
	}

	return dto.CreateUserResp{
		UserID:       userID,
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
	}, nil
}
