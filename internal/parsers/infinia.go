package parsers

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/hackrush01/cardsplit/internal/models"
)

const separator = "~|~"

// ParseInfiniaCSV reads the raw HDFC Infinia file and extracts a full Statement DTO.
func ParseInfiniaCSV(r io.Reader) (*models.Statement, error) {
	scanner := bufio.NewScanner(r)

	var headerLines, txnLines, rewardLines []string
	currentSegment := "Header"

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Detect segment transitions
		switch {
		case strings.HasPrefix(line, "Domestic / International Transactions"):
			currentSegment = "Transactions"
			continue
		case strings.HasPrefix(line, "Reward Points Summary"):
			currentSegment = "Rewards"
			continue
		}

		// Group lines by segment
		switch currentSegment {
		case "Header":
			headerLines = append(headerLines, line)
		case "Transactions":
			txnLines = append(txnLines, line)
		case "Rewards":
			rewardLines = append(rewardLines, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	stmt := &models.Statement{CardType: "Infinia"}
	parseHeader(headerLines, stmt)
	transactions, warnings := parseTransactions(txnLines)
	stmt.Transactions = transactions
	stmt.Warnings = warnings

	return stmt, nil
}

func parseHeader(lines []string, stmt *models.Statement) {
	for _, line := range lines {
		cols := trimmedSplit(line, separator)
		if len(cols) < 2 {
			continue
		}
		key, val := cols[0], cols[1]

		switch key {
		case "Statement Date":
			stmt.StatementDate = parseDate(val)
		case "Payment Due Date":
			stmt.PaymentDueDate = parseDate(val)
		case "Total Amount Due":
			stmt.TotalAmountDue = parseAmount(val)
		case "Minimum Amount Due":
			stmt.MinAmountDue = parseAmount(val)
		}
	}
}

func parseTransactions(lines []string) ([]models.Transaction, []string) {
	var txns []models.Transaction
	var warnings []string
	if len(lines) < 2 { // Need at least the table header + 1 data row
		return txns, warnings
	}

	seen := map[int64]struct{}{} // Use a map to track seen timestamps for duplicate detection

	// Skip the header row (Transaction type~|~...)
	for _, line := range lines[1:] {
		cols := trimmedSplit(line, separator)
		if len(cols) < 6 {
			warnings = append(warnings, fmt.Sprintf("skipping malformed transaction line: %s", line))
			continue
		}

		txnTime := parseDate(cols[2])
		if txnTime.IsZero() {
			warnings = append(warnings, fmt.Sprintf("invalid date format for transaction: %s", cols[2]))
			continue
		}

		keyTime := txnTime
		if _, exists := seen[keyTime.Unix()]; exists {
			o := keyTime.Format("02/01/2006 15:04:05")
			for {
				keyTime = keyTime.Add(time.Second)
				if _, exists := seen[keyTime.Unix()]; !exists {
					break
				}
			}
			warnings = append(warnings, fmt.Sprintf("duplicate timestamp found for %s; shifted to %s", o, keyTime.Format("02/01/2006 15:04:05")))
		}
		seen[keyTime.Unix()] = struct{}{}

		amount := parseAmount(cols[4])
		if cols[5] == "Cr" {
			amount = -amount
		}

		var brv int
		if len(cols) >= 7 && cols[6] != "" {
			val := strings.ReplaceAll(cols[6], " ", "")
			brv, _ = strconv.Atoi(val)
		}

		txns = append(txns, models.Transaction{
			Type:             cols[0],
			CardHolderName:   cols[1],
			TxnTimestamp:     txnTime,
			KeyTimestamp:     keyTime,
			Description:      cols[3],
			Amount:           amount,
			BaseRewardValue:  brv,
			RewardMultiplier: 1,
		})
	}
	return txns, warnings
}

func parseDate(s string) time.Time {
	formats := []string{"02/01/2006 15:04:05", "02/01/2006"}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func parseAmount(s string) int {
	s = strings.ReplaceAll(strings.TrimSpace(s), ",", "")
	s = strings.ReplaceAll(strings.TrimSpace(s), ".", "")
	if s == "" || s == "-" {
		return 0
	}

	// Convert to paise to avoid float issues (e.g. 123.45 becomes 12345)
	paise, _ := strconv.Atoi(s)
	return paise
}

func trimmedSplit(line, sep string) []string {
	parts := strings.Split(line, sep)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
