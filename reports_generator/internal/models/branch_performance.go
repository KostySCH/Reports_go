package models

type BranchInfo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Location    string `json:"location"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	ManagerName string `json:"manager_name"`
}

type CustomerStats struct {
	TotalCustomers int `json:"total_customers"`
	TotalAccounts  int `json:"total_accounts"`
	ActiveAccounts int `json:"active_accounts"`
}

type TransactionStats struct {
	TotalTransactions int     `json:"total_transactions"`
	TotalAmount       float64 `json:"total_amount"`
	AverageAmount     float64 `json:"average_amount"`
}

type DailyActivity struct {
	Date          string  `json:"date"`
	Transactions  int     `json:"transactions"`
	Amount        float64 `json:"amount"`
	GrowthPercent float64 `json:"growth_percent"`
}

type TopCustomer struct {
	Name         string  `json:"name"`
	Transactions int     `json:"transactions"`
	TotalAmount  float64 `json:"total_amount"`
}

type BranchPerformanceData struct {
	BranchInfo       BranchInfo       `json:"branch_info"`
	CustomerStats    CustomerStats    `json:"customer_stats"`
	TransactionStats TransactionStats `json:"transaction_stats"`
	DailyActivity    []DailyActivity  `json:"daily_activity"`
	TopCustomers     []TopCustomer    `json:"top_customers"`
}
