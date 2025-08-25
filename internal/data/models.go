package data

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

func Handlectx() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	return ctx, cancel
}

type Models struct {
	Tokens TokensModel
}

func NewModel(db *sqlx.DB) Models {
	return Models{
		Tokens: TokensModel{DB: db},
	}
}
