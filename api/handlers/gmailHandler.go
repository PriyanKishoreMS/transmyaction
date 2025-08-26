package handlers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/priyankishorems/transmyaction/utils"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

func GmailService(token *oauth2.Token) (*gmail.Service, error) {
	client := oauthConfig.Client(context.Background(), token)
	srv, err := gmail.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func (h *Handlers) updateTokens(token *oauth2.Token, email string) (*oauth2.Token, error) {

	ts := oauthConfig.TokenSource(context.Background(), token)

	refreshedToken, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	if refreshedToken.AccessToken != token.AccessToken || refreshedToken.Expiry != token.Expiry {
		err := h.Data.Tokens.UpdateTokens(
			refreshedToken,
			email,
		)
		if err != nil {
			return nil, err
		}
	}

	return refreshedToken, nil
}

func (h *Handlers) GmailHandler(c echo.Context) error {

	email := c.Param("email")

	token, err := h.Data.Tokens.GetTokenFromEmail(email)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("can't get token from db: %v", err)
	}

	refreshedToken, err := h.updateTokens(token, email)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("can't refresh token: %v", err)
	}

	token = refreshedToken

	srv, err := GmailService(token)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("can't get gmail service: %v", err)
	}

	allTxns, err := getAllMails(srv, email, "alerts@axisbank.com", "2025/07/01")
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("can't get mails: %v", err)
	}

	err = h.Data.Txns.SaveTransactions(allTxns)
	if err != nil {
		h.Utils.InternalServerError(c, err)
		return fmt.Errorf("can't add mails in db: %v", err)
	}

	return c.JSON(http.StatusOK, Cake{
		"message": "wokring",
	})
}

func getAllMails(srv *gmail.Service, userEmail string, mailID string, afterDate string) ([]utils.Transaction, error) {

	call := srv.Users.Messages.List("me").
		Q(fmt.Sprintf("from:%s after:%s", mailID, afterDate)).MaxResults(500)

	msgs, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("can't get mails: %v", err)

	}

	var allTxn []utils.Transaction

	for i, m := range msgs.Messages {
		full, err := srv.Users.Messages.Get("me", m.Id).Format("full").Do()
		if err != nil {
			return nil, fmt.Errorf("can't fetch message %s: %v", m.Id, err)
		}

		fmt.Println("Fetched email no: ", i+1)

		body, err := extractBody(full)
		if err != nil {
			return nil, fmt.Errorf("can't extract body: %v", err)
		}

		fmt.Println("Extracted email no: ", i+1)

		txn, err := parseAxisEmail(body, userEmail)

		for _, h := range full.Payload.Headers {
			if h.Name == "Date" {
				t, _ := time.Parse(time.RFC1123Z, h.Value) // sometimes RFC1123 or RFC822
				txn.TxnDatetime = t
			}
		}

		allTxn = append(allTxn, *txn)

		if err != nil {
			// not all mails will be valid transactions, just skip
			continue
		}

		fmt.Printf("Transaction: %v", txn)
	}
	return allTxn, nil
}

func extractBody(msg *gmail.Message) (string, error) {
	if msg.Payload == nil {
		return "", fmt.Errorf("no payload")
	}

	var data string
	if msg.Payload.Body != nil && msg.Payload.Body.Data != "" {
		data = msg.Payload.Body.Data
	} else if len(msg.Payload.Parts) > 0 {
		for _, p := range msg.Payload.Parts {
			if p.MimeType == "text/plain" || p.MimeType == "text/html" {
				data = p.Body.Data
				break
			}
		}
	}

	if data == "" {
		return "", fmt.Errorf("no body data")
	}

	decoded, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

func parseAxisEmail(body, userEmail string) (*utils.Transaction, error) {
	var txn utils.Transaction
	txn.UserEmail = userEmail

	// amount
	reAmount := regexp.MustCompile(`INR\s+([\d,.]+)`)
	if match := reAmount.FindStringSubmatch(body); match != nil {
		amt, _ := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", ""), 64)
		txn.Amount = amt
	}

	// debit/credit
	if strings.Contains(strings.ToLower(body), "debited") {
		txn.TxnType = "debit"
	} else if strings.Contains(strings.ToLower(body), "credited") {
		txn.TxnType = "credit"
	}

	// account
	reAcc := regexp.MustCompile(`(?:A/c no\.|Account Number:)\s*([X\d]+)`)
	if match := reAcc.FindStringSubmatch(body); match != nil {
		txn.AccountNumber = match[1]
	}

	// datetime
	// reDate := regexp.MustCompile(`(\d{2}-\d{2}-\d{4}),\s*(\d{2}:\d{2}:\d{2})`)
	// if match := reDate.FindStringSubmatch(body); match != nil {
	// 	layout := "02-01-2006, 15:04:05"
	// 	t, _ := time.Parse(layout, match[1]+", "+match[2])
	// 	txn.TxnDatetime = t
	// }

	// txn info
	reTxnInfo := regexp.MustCompile(`([A-Z0-9\.]+(?:/[A-Z0-9\.]+)+/(?:[A-Z0-9 ]+))`)
	if match := reTxnInfo.FindStringSubmatch(body); match != nil {
		txn.TxnInfo = strings.TrimSpace(match[1]) // full string

		parts := strings.Split(txn.TxnInfo, "/")

		switch parts[0] {
		case "UPI":
			// Expect: UPI / P2A / ref / counterparty
			if len(parts) >= 2 {
				txn.TxnMode = parts[1]
			}
			if len(parts) >= 3 {
				txn.TxnRef = parts[2]
			}
			if len(parts) >= 4 {
				txn.CounterParty = strings.TrimSpace(parts[3])
			}
			txn.TxnMethod = "UPI"

		case "NEFT":
			// Expect: NEFT / ref / counterparty
			if len(parts) >= 2 {
				txn.TxnRef = parts[1]
			}
			if len(parts) >= 3 {
				txn.CounterParty = strings.TrimSpace(parts[2])
			}
			txn.TxnMethod = "NEFT"

		default:
			// Fallback — just record whatever we can
			txn.TxnMethod = parts[0]
			if len(parts) > 1 {
				txn.TxnRef = parts[1]
			}
			if len(parts) > 2 {
				txn.CounterParty = strings.TrimSpace(parts[len(parts)-1])
			}
		}
	}

	return &txn, nil
}

// Q(`from:alerts@axisbank.com after:2025/02/25`).

// func getAllMails(srv *gmail.Service, userEmail string, mailID string, afterDate string) error {

// 	err := srv.Users.Messages.List("me").
// 		Q(fmt.Sprintf("from:%s after:%s", mailID, afterDate)).Pages(context.TODO(), func(msgs *gmail.ListMessagesResponse) error {
// 		for i, m := range msgs.Messages {
// 			full, err := srv.Users.Messages.Get("me", m.Id).Format("full").Do()
// 			if err != nil {
// 				return fmt.Errorf("can't fetch message %s: %v", m.Id, err)
// 			}

// 			fmt.Println("Fetched email no: ", i+1)

// 			body, err := extractBody(full)
// 			if err != nil {
// 				return fmt.Errorf("can't extract body: %v", err)
// 			}

// 			fmt.Println("Extracted email no: ", i+1)

// 			txn, err := parseAxisEmail(body, userEmail)

// 			for _, h := range full.Payload.Headers {
// 				if h.Name == "Date" {
// 					t, _ := time.Parse(time.RFC1123Z, h.Value) // sometimes RFC1123 or RFC822
// 					txn.TxnDatetime = t
// 				}
// 			}

// 			if err != nil {
// 				// not all mails will be valid transactions, just skip
// 				continue
// 			}

// 			fmt.Printf("Transaction: %v", txn)
// 		}
// 		return nil
// 	})

// 	if err != nil {
// 		return fmt.Errorf("can't get mails: %v", err)

// 	}

// 	return nil
// }
