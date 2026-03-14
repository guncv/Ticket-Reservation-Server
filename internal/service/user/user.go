package user

import (
	"context"

	"github.com/guncv/ticket-reservation-server/internal/service/user/dto"
	"github.com/guncv/ticket-reservation-server/internal/service/user/repo"
	"github.com/guncv/ticket-reservation-server/internal/service/user/user"
	"github.com/guncv/ticket-reservation-server/internal/shared"
)

func (s *userService) createUser(ctx context.Context, req dto.CreateUserReq, role dto.UserRole) (dto.CreateUserResp, error) {
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
		Role:           role,
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
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
	}, nil
}

func (s *userService) CreateUser(ctx context.Context, req dto.CreateUserReq) (dto.CreateUserResp, error) {
	return s.createUser(ctx, req, dto.UserRoleUser)
}

func (s *userService) CreateAdminUser(ctx context.Context, req dto.CreateUserReq) (dto.CreateUserResp, error) {
	return s.createUser(ctx, req, dto.UserRoleAdmin)
}

func (s *userService) LoginUser(ctx context.Context, req dto.LoginUserReq) (dto.LoginUserResp, error) {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return dto.LoginUserResp{}, err
	}
	defer tx.Rollback(ctx)

	usr, err := s.userRepo.GetUserByUserName(ctx, req.UserName)
	if err != nil {
		return dto.LoginUserResp{}, err
	}

	err = user.ComparePassword(usr.HashedPassword, req.Password)
	if err != nil {
		return dto.LoginUserResp{}, err
	}

	session, err := s.CreateSession(ctx, usr.ID)
	if err != nil {
		return dto.LoginUserResp{}, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return dto.LoginUserResp{}, err
	}

	return dto.LoginUserResp{
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
	}, nil
}

func (s *userService) LogoutUser(ctx context.Context) error {
	ctx, tx, err := s.db.EnsureTxFromCtx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	refreshToken, err := shared.GetRefreshTokenFromContext(ctx)
	if err != nil {
		return err
	}

	session, err := s.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return err
	}

	err = s.userRepo.RevokeSession(ctx, session.ID)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
