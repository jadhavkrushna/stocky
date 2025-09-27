package services

import (
	"database/sql"
	"fmt"

	"stocky-backend/internal/models"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

type LedgerService struct {
	db     *sql.DB
	logger *logrus.Logger
}

func NewLedgerService(db *sql.DB, logger *logrus.Logger) *LedgerService {
	return &LedgerService{
		db:     db,
		logger: logger,
	}
}

func (s *LedgerService) GetLedgerEntries(rewardID uuid.UUID) ([]models.LedgerEntry, error) {
	rows, err := s.db.Query(`
		SELECT id, reward_id, adjustment_id, entry_type, account_type, 
		       account_identifier, amount_inr, quantity, stock_id, description, created_at
		FROM ledger_entries
		WHERE reward_id = $1
		ORDER BY created_at`, rewardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger entries: %w", err)
	}
	defer rows.Close()
	
	var entries []models.LedgerEntry
	for rows.Next() {
		var entry models.LedgerEntry
		if err := rows.Scan(&entry.ID, &entry.RewardID, &entry.AdjustmentID,
			&entry.EntryType, &entry.AccountType, &entry.AccountIdentifier,
			&entry.AmountINR, &entry.Quantity, &entry.StockID,
			&entry.Description, &entry.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan ledger entry: %w", err)
		}
		entries = append(entries, entry)
	}
	
	return entries, nil
}

func (s *LedgerService) GetAccountBalance(accountType, accountIdentifier string) (decimal.Decimal, error) {
	var balance decimal.Decimal
	
	err := s.db.QueryRow(`
		SELECT COALESCE(
			SUM(CASE WHEN entry_type = 'CREDIT' THEN amount_inr ELSE -amount_inr END), 
			0
		)
		FROM ledger_entries
		WHERE account_type = $1 AND account_identifier = $2`,
		accountType, accountIdentifier).Scan(&balance)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get account balance: %w", err)
	}
	
	return balance, nil
}

func (s *LedgerService) GetUserStockBalance(userID uuid.UUID, stockID uuid.UUID) (decimal.Decimal, error) {
	var balance decimal.Decimal
	
	err := s.db.QueryRow(`
		SELECT COALESCE(
			SUM(CASE WHEN entry_type = 'CREDIT' THEN quantity ELSE -quantity END), 
			0
		)
		FROM ledger_entries
		WHERE account_type = 'USER_STOCK_HOLDINGS' 
		AND account_identifier = $1 
		AND stock_id = $2`,
		userID.String(), stockID).Scan(&balance)
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to get user stock balance: %w", err)
	}
	
	return balance, nil
}

func (s *LedgerService) GetCompanyExpenses() (map[string]decimal.Decimal, error) {
	rows, err := s.db.Query(`
		SELECT account_type, SUM(amount_inr) as total
		FROM ledger_entries
		WHERE entry_type = 'DEBIT' AND account_type LIKE 'COMPANY_%'
		GROUP BY account_type`)
	if err != nil {
		return nil, fmt.Errorf("failed to get company expenses: %w", err)
	}
	defer rows.Close()
	
	expenses := make(map[string]decimal.Decimal)
	for rows.Next() {
		var accountType string
		var total decimal.Decimal
		if err := rows.Scan(&accountType, &total); err != nil {
			return nil, fmt.Errorf("failed to scan expense: %w", err)
		}
		expenses[accountType] = total
	}
	
	return expenses, nil
}

func (s *LedgerService) VerifyLedgerBalance() error {
	// Verify that total debits equal total credits
	var totalDebits, totalCredits decimal.Decimal
	
	err := s.db.QueryRow(`
		SELECT 
			COALESCE(SUM(CASE WHEN entry_type = 'DEBIT' THEN amount_inr END), 0),
			COALESCE(SUM(CASE WHEN entry_type = 'CREDIT' THEN amount_inr END), 0)
		FROM ledger_entries`).Scan(&totalDebits, &totalCredits)
	if err != nil {
		return fmt.Errorf("failed to get ledger totals: %w", err)
	}
	
	if !totalDebits.Equal(totalCredits) {
		return fmt.Errorf("ledger is out of balance: debits=%s, credits=%s", 
			totalDebits.String(), totalCredits.String())
	}
	
	s.logger.WithFields(logrus.Fields{
		"total_debits":  totalDebits.String(),
		"total_credits": totalCredits.String(),
	}).Info("Ledger balance verification passed")
	
	return nil
}

func (s *LedgerService) GetLedgerReport() (map[string]interface{}, error) {
	report := make(map[string]interface{})
	
	// Get total debits and credits
	var totalDebits, totalCredits decimal.Decimal
	err := s.db.QueryRow(`
		SELECT 
			COALESCE(SUM(CASE WHEN entry_type = 'DEBIT' THEN amount_inr END), 0),
			COALESCE(SUM(CASE WHEN entry_type = 'CREDIT' THEN amount_inr END), 0)
		FROM ledger_entries`).Scan(&totalDebits, &totalCredits)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledger totals: %w", err)
	}
	
	report["total_debits"] = totalDebits.String()
	report["total_credits"] = totalCredits.String()
	report["is_balanced"] = totalDebits.Equal(totalCredits)
	
	// Get balances by account type
	rows, err := s.db.Query(`
		SELECT 
			account_type,
			entry_type,
			SUM(amount_inr) as total
		FROM ledger_entries
		GROUP BY account_type, entry_type
		ORDER BY account_type, entry_type`)
	if err != nil {
		return nil, fmt.Errorf("failed to get account balances: %w", err)
	}
	defer rows.Close()
	
	accountBalances := make(map[string]map[string]string)
	for rows.Next() {
		var accountType, entryType string
		var total decimal.Decimal
		if err := rows.Scan(&accountType, &entryType, &total); err != nil {
			return nil, fmt.Errorf("failed to scan account balance: %w", err)
		}
		
		if accountBalances[accountType] == nil {
			accountBalances[accountType] = make(map[string]string)
		}
		accountBalances[accountType][entryType] = total.String()
	}
	
	report["account_balances"] = accountBalances
	
	// Get company expenses summary
	expenses, err := s.GetCompanyExpenses()
	if err != nil {
		return nil, fmt.Errorf("failed to get company expenses: %w", err)
	}
	
	expenseStrings := make(map[string]string)
	for k, v := range expenses {
		expenseStrings[k] = v.String()
	}
	report["company_expenses"] = expenseStrings
	
	return report, nil
}