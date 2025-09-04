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

func (h *Handlers) UpdateTransactionsHandler(c echo.Context) error {

	email := c.Param("email")

	requestID := fmt.Sprintf("REQ-%d", time.Now().UnixNano())

	c.Response().Header().Set("Content-Type", "application/json")
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Write([]byte(fmt.Sprintf(`{"message":"Processing started","request_id":"%s","status":"in_progress"}`, requestID)))
	c.Response().Flush()

	fmt.Printf("BROWSER RESPONSE SENT - Processing continues in background\n")

	go func() {

		token, err := h.Data.Tokens.GetTokenFromEmail(email)
		if err != nil {
			h.Utils.InternalServerError(c, err)
			return
		}

		refreshedToken, err := h.updateTokens(token, email)
		if err != nil {
			h.Utils.InternalServerError(c, err)
			return
		}

		token = refreshedToken

		srv, err := GmailService(token)
		if err != nil {
			h.Utils.InternalServerError(c, err)
			return
		}

		allTxns, err := postAllMails(srv, email, "alerts@axisbank.com", "2025/07/01")
		if err != nil {
			h.Utils.InternalServerError(c, err)
			return
		}

		err = h.Data.Txns.SaveTransactions(allTxns)
		if err != nil {
			h.Utils.InternalServerError(c, err)
			return
		}

	}()

	return nil
}

func FormatTimeForAxis(t time.Time) string {

	t, err := time.Parse(time.RFC3339, t.Format(time.RFC3339))
	if err != nil {
		panic(err)
	}

	// Format into yyyy/mm/dd
	return t.Format("2006/01/02")
}

func (h *Handlers) UpdateTransactionsJob() error {

	distinctEmails, err := h.Data.Txns.GetAllDistinctEmails()
	if err != nil {
		return nil
	}

	for _, user := range distinctEmails {

		fmt.Println("Getting token for user:", user.Email)
		token, err := h.Data.Tokens.GetTokenFromEmail(user.Email)
		if err != nil {
			fmt.Println("Error getting token:", err)
			return nil
		}

		fmt.Println("Refreshing token for user:", user.Email)
		refreshedToken, err := h.updateTokens(token, user.Email)
		if err != nil {
			return nil
		}

		token = refreshedToken

		srv, err := GmailService(token)
		if err != nil {
			return nil
		}

		afterDate := FormatTimeForAxis(user.LastUpdated)

		fmt.Println("Fetching mails for user:", user.Email, "after date:", afterDate)
		allTxns, err := postAllMails(srv, user.Email, "alerts@axisbank.com", afterDate)
		if err != nil {
			return nil
		}

		fmt.Println("Saving transactions for user:", user.Email)
		err = h.Data.Txns.SaveTransactions(allTxns)
		if err != nil {
			return nil
		}
		fmt.Println("Saved transactions for user:", user.Email, "count:", len(allTxns))
		fmt.Println("-----------------------------------------------------")
	}

	return nil
}

func postAllMails(srv *gmail.Service, userEmail string, mailID string, afterDate string) ([]utils.Transaction, error) {
	var allTxn []utils.Transaction

	call := srv.Users.Messages.List("me").
		Q(fmt.Sprintf("from:%s after:%s before:%s", mailID, afterDate, time.Now().Add(24*time.Hour).Format("2006/01/02"))).MaxResults(500)

	msgs, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("can't get mails: %v", err)

	}

	fmt.Println("length of msgs bro: ", len(msgs.Messages))

	for i, m := range msgs.Messages {
		full, err := srv.Users.Messages.Get("me", m.Id).Format("full").Do()
		if err != nil {
			return nil, fmt.Errorf("can't fetch message %s: %v", m.Id, err)
		}

		fmt.Println("Fetched email no: ", i+1)

		body, err := extractBody(full)
		if err != nil {
			fmt.Printf("can't extract body: %v", err)
		}

		fmt.Println("Extracted email no: ", i+1)

		txn, err := parseAxisEmail(body, userEmail)
		if err != nil {
			// not all mails will be valid transactions, just skip
			continue
		}

		for _, h := range full.Payload.Headers {
			if h.Name == "Date" {
				t, _ := time.Parse(time.RFC1123Z, h.Value) // sometimes RFC1123 or RFC822
				txn.TxnDatetime = t
			}
		}

		allTxn = append(allTxn, *txn)

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
	reTxnInfo := regexp.MustCompile(`([A-Za-z0-9\.]+(?:/[A-Za-z0-9\.]+)+/(?:[A-Za-z0-9 ]+))`)
	if match := reTxnInfo.FindStringSubmatch(body); match != nil {
		txn.TxnInfo = strings.TrimSpace(match[1]) // full string

		parts := strings.Split(txn.TxnInfo, "/")

		switch len(parts) {
		case 4:
			// Expect: UPI / P2A / ref / counterparty
			txn.TxnMode = parts[1]
			txn.TxnRef = parts[2]
			txn.CounterParty = strings.TrimSpace(parts[3])
			txn.TxnMethod = parts[0]

		case 3:
			// Expect: NEFT / ref / counterparty
			txn.TxnRef = parts[1]
			txn.CounterParty = strings.TrimSpace(parts[2])
			txn.TxnMethod = parts[0]

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

func (h *Handlers) GetTransactionsHandler(c echo.Context) error {
	email := c.Param("email")
	interval := c.Param("interval") // "3m", "6m", "1y", etc

	year, _ := strconv.Atoi(c.Param("year"))
	month, _ := strconv.Atoi(c.Param("month"))

	fromStr := c.QueryParam("from") // YYYY-MM-DD
	toStr := c.QueryParam("to")

	var from, to *time.Time
	if fromStr != "" && toStr != "" {
		f, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid from date"})
		}
		t, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid to date"})
		}
		from, to = &f, &t
	}

	txns, err := h.Data.Txns.GetTransactions(email, interval, year, month, from, to)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, txns)
}
