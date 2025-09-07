package data

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pascaldekloe/jwt"
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

	fmt.Println("saving token for ", email, name, token)

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

func (t TokensModel) GenerateAuthTokens(id string, secret string, issuer string) ([]byte, []byte, error) {
	byteSecret := []byte(secret)
	accessToken, err := GenerateAccessToken(id, byteSecret, issuer)
	if err != nil {
		return nil, nil, err
	}
	refreshToken, err := GenerateRefreshToken(id, byteSecret, issuer)
	if err != nil {
		return nil, nil, err
	}

	return accessToken, refreshToken, nil
}

func GenerateAccessToken(id string, secret []byte, issuer string) ([]byte, error) {
	var claims jwt.Claims
	claims.Subject = id
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add(time.Hour * 36))
	claims.Issuer = issuer
	claims.Set = map[string]interface{}{
		"type": "access",
	}

	accessToken, err := claims.HMACSign(jwt.HS256, secret)
	if err != nil {
		return nil, err
	}

	return accessToken, nil
}

func GenerateRefreshToken(id string, secret []byte, issuer string) ([]byte, error) {
	var claims jwt.Claims
	claims.Subject = id
	claims.Issued = jwt.NewNumericTime(time.Now())
	claims.Expires = jwt.NewNumericTime(time.Now().Add((time.Hour * 24) * 90))
	claims.Issuer = issuer
	claims.Set = map[string]interface{}{
		"type": "refresh",
	}

	refreshToken, err := claims.HMACSign(jwt.HS256, secret)
	if err != nil {
		return nil, err
	}

	return refreshToken, nil
}
