package types

import (
	"testing"
	"time"
)

func TestFundValidation(t *testing.T) {
	tests := []struct {
		name    string
		fund    Fund
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid fund",
			fund: Fund{
				TenantID:   1,
				MerchantID: 1,
				FundType:   FundTypeDeposit,
				Amount:     100.0,
				Currency:   "CNY",
				Status:     FundStatusPending,
			},
			wantErr: false,
		},
		{
			name: "missing tenant_id",
			fund: Fund{
				MerchantID: 1,
				FundType:   FundTypeDeposit,
				Amount:     100.0,
				Currency:   "CNY",
				Status:     FundStatusPending,
			},
			wantErr: true,
			errMsg:  "租户ID不能为空",
		},
		{
			name: "missing merchant_id",
			fund: Fund{
				TenantID: 1,
				FundType: FundTypeDeposit,
				Amount:   100.0,
				Currency: "CNY",
				Status:   FundStatusPending,
			},
			wantErr: true,
			errMsg:  "商户ID不能为空",
		},
		{
			name: "invalid amount",
			fund: Fund{
				TenantID:   1,
				MerchantID: 1,
				FundType:   FundTypeDeposit,
				Amount:     -100.0,
				Currency:   "CNY",
				Status:     FundStatusPending,
			},
			wantErr: true,
			errMsg:  "金额必须大于0",
		},
		{
			name: "invalid currency",
			fund: Fund{
				TenantID:   1,
				MerchantID: 1,
				FundType:   FundTypeDeposit,
				Amount:     100.0,
				Currency:   "INVALID",
				Status:     FundStatusPending,
			},
			wantErr: true,
			errMsg:  "货币代码必须为3位",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fund.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestFundTransactionValidation(t *testing.T) {
	tests := []struct {
		name        string
		transaction FundTransaction
		wantErr     bool
		errMsg      string
	}{
		{
			name: "valid credit transaction",
			transaction: FundTransaction{
				TenantID:        1,
				MerchantID:      1,
				FundID:          1,
				TransactionType: TransactionTypeCredit,
				Amount:          100.0,
				BalanceBefore:   200.0,
				BalanceAfter:    300.0,
				OperatorID:      1,
			},
			wantErr: false,
		},
		{
			name: "valid debit transaction",
			transaction: FundTransaction{
				TenantID:        1,
				MerchantID:      1,
				FundID:          1,
				TransactionType: TransactionTypeDebit,
				Amount:          100.0,
				BalanceBefore:   300.0,
				BalanceAfter:    200.0,
				OperatorID:      1,
			},
			wantErr: false,
		},
		{
			name: "invalid credit calculation",
			transaction: FundTransaction{
				TenantID:        1,
				MerchantID:      1,
				FundID:          1,
				TransactionType: TransactionTypeCredit,
				Amount:          100.0,
				BalanceBefore:   200.0,
				BalanceAfter:    250.0, // 应该是300.0
				OperatorID:      1,
			},
			wantErr: true,
			errMsg:  "入账余额计算错误",
		},
		{
			name: "invalid debit calculation",
			transaction: FundTransaction{
				TenantID:        1,
				MerchantID:      1,
				FundID:          1,
				TransactionType: TransactionTypeDebit,
				Amount:          100.0,
				BalanceBefore:   300.0,
				BalanceAfter:    250.0, // 应该是200.0
				OperatorID:      1,
			},
			wantErr: true,
			errMsg:  "出账余额计算错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.transaction.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestDepositRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request DepositRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid deposit request",
			request: DepositRequest{
				MerchantID: 1,
				Amount:     100.0,
				Currency:   "CNY",
			},
			wantErr: false,
		},
		{
			name: "amount too large",
			request: DepositRequest{
				MerchantID: 1,
				Amount:     2000000.0, // 超过限制
				Currency:   "CNY",
			},
			wantErr: true,
			errMsg:  "单笔充值金额不能超过1,000,000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
			}
		})
	}
}

func TestRightsBalanceUpdate(t *testing.T) {
	balance := &RightsBalance{
		TotalBalance:  1000.0,
		UsedBalance:   300.0,
		FrozenBalance: 100.0,
	}

	balance.UpdateAvailableBalance()

	expectedAvailable := 600.0 // 1000 - 300 - 100
	if balance.AvailableBalance != expectedAvailable {
		t.Errorf("Expected AvailableBalance %f, got %f", expectedAvailable, balance.AvailableBalance)
	}

	if balance.LastUpdated.IsZero() {
		t.Error("LastUpdated should be set")
	}

	if time.Since(balance.LastUpdated) > time.Second {
		t.Error("LastUpdated should be recent")
	}
}

func TestFundTypeString(t *testing.T) {
	tests := []struct {
		fundType FundType
		expected string
	}{
		{FundTypeDeposit, "deposit"},
		{FundTypeAllocation, "allocation"},
		{FundTypeConsumption, "consumption"},
		{FundTypeRefund, "refund"},
		{FundType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.fundType.String(); got != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestTransactionTypeString(t *testing.T) {
	tests := []struct {
		transType TransactionType
		expected  string
	}{
		{TransactionTypeCredit, "credit"},
		{TransactionTypeDebit, "debit"},
		{TransactionType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.transType.String(); got != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, got)
			}
		})
	}
}