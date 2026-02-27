package app

type RegisterParams struct {
	Username string `json:"username" validate:"required"`
}

type RegisterResult struct {
	ID int64 `json:"id"`
}

func (ctx *Context) Register(params RegisterParams) (*RegisterResult, error) {
	logger := ctx.getLogger()

	if err := validateInput(params); err != nil {
		logger.Errorf("validateInput error : %s", err)
		return nil, err
	}

	userID, err := ctx.DB.CreateUser(params.Username)
	if err != nil {
		return nil, err
	}

	return &RegisterResult{
		ID: userID,
	}, nil
}
