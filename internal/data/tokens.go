package data

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/oauth2"
)

type TokensModel struct {
	DB *sqlx.DB
}

type Token struct {
	AccessToken  string    `db:"access_token"`
	RefreshToken string    `db:"refresh_token"`
	TokenType    string    `db:"token_type"`
	Expiry       time.Time `db:"expiry"`
}

func (t TokensModel) SaveEmailToken(email string, name string, token *oauth2.Token) error {
	ctx, cancel := Handlectx()
	defer cancel()

	query := `
	INSERT INTO gmail_tokens (
  	user_email, user_name, access_token, refresh_token, token_type, expiry, updated_at
		) VALUES (
  		?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP
		)
	ON CONFLICT(user_email) DO UPDATE SET
  	access_token  = excluded.access_token,
  	token_type    = excluded.token_type,
  	expiry        = excluded.expiry,
  	updated_at    = CURRENT_TIMESTAMP,
	refresh_token = excluded.refresh_token
`

	_, err := t.DB.ExecContext(ctx, query, email, name, token.AccessToken, token.RefreshToken, token.TokenType, token.Expiry)
	if err != nil {
		return err
	}

	return nil
}

func (t TokensModel) GetTokenFromEmail(email string) (*oauth2.Token, error) {
	ctx, cancel := Handlectx()
	defer cancel()

	query := `
	SELECT access_token, refresh_token, token_type, expiry
    FROM gmail_tokens
    WHERE user_email = ?
	`

	var tok Token

	if err := t.DB.GetContext(ctx, &tok, query, email); err != nil {
		return nil, fmt.Errorf("can't get token from email: %v", err)
	}

	token := &oauth2.Token{
		AccessToken:  tok.AccessToken,
		RefreshToken: tok.RefreshToken,
		TokenType:    tok.TokenType,
		Expiry:       tok.Expiry,
	}

	return token, nil
}

func (t TokensModel) UpdateTokens(refreshed *oauth2.Token, email string) error {
	ctx, cancel := Handlectx()
	defer cancel()

	fmt.Println("test nigga", refreshed.AccessToken, refreshed.Expiry, refreshed.RefreshToken, refreshed.TokenType, email, "end nigga")

	_, err := t.DB.ExecContext(ctx, `
			UPDATE gmail_tokens
			SET access_token = ?, expiry = ?, refresh_token = ?, token_type = ?, updated_at = CURRENT_TIMESTAMP
			WHERE user_email = ?
		`, refreshed.AccessToken, refreshed.Expiry, refreshed.RefreshToken, refreshed.TokenType, email)
	if err != nil {
		return fmt.Errorf("error in updating token: %v", err)
	}
	return nil
}
