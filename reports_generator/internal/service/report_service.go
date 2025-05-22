package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/KostySCH/Reports_go/reports_generator/internal/models"
)

type ReportService struct {
	db         *sql.DB
	docService *DocumentService
	minioSvc   *MinioService
}

func NewReportService(db *sql.DB, outputDir string, minioSvc *MinioService) *ReportService {
	return &ReportService{
		db:         db,
		docService: NewDocumentService(outputDir),
		minioSvc:   minioSvc,
	}
}

func (s *ReportService) GenerateBranchPerformanceReport(ctx context.Context, params *models.BranchPerformanceParams) (string, error) {

	reportPath, err := s.generateReport(ctx, params)
	if err != nil {
		return "", err
	}

	minioPath, err := s.minioSvc.UploadReport(ctx, reportPath, fmt.Sprintf("%d", params.BranchID))
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки отчета в MinIO: %v", err)
	}

	return minioPath, nil
}

func (s *ReportService) generateReport(ctx context.Context, params *models.BranchPerformanceParams) (string, error) {

	fullDate := params.Month + "-01"

	branchInfo, err := s.getBranchInfo(ctx, params.BranchID)
	if err != nil {
		return "", fmt.Errorf("ошибка получения информации о филиале: %v", err)
	}

	// Получаем статистику клиентов
	customerStats, err := s.getCustomerStats(ctx, params.BranchID)
	if err != nil {
		return "", fmt.Errorf("ошибка получения статистики клиентов: %v", err)
	}

	// Получаем статистику транзакций
	transactionStats, err := s.getTransactionStats(ctx, params.BranchID, fullDate)
	if err != nil {
		return "", fmt.Errorf("ошибка получения статистики транзакций: %v", err)
	}

	// Получаем ежедневную активность
	dailyActivity, err := s.getDailyActivity(ctx, params.BranchID, fullDate)
	if err != nil {
		return "", fmt.Errorf("ошибка получения ежедневной активности: %v", err)
	}

	// Получаем топ клиентов
	topCustomers, err := s.getTopCustomers(ctx, params.BranchID, fullDate)
	if err != nil {
		return "", fmt.Errorf("ошибка получения топ клиентов: %v", err)
	}

	// Формируем данные для отчета
	data := &models.BranchPerformanceData{
		BranchInfo:       *branchInfo,
		CustomerStats:    *customerStats,
		TransactionStats: *transactionStats,
		DailyActivity:    dailyActivity,
		TopCustomers:     topCustomers,
	}

	// Генерируем отчет
	return s.docService.GenerateReport(data, params.Format)
}

func (s *ReportService) getBranchInfo(ctx context.Context, branchID int64) (*models.BranchInfo, error) {
	query := `
		SELECT b.branch_id, b.branch_name, b.location, b.phone, b.email, 
		       e.first_name || ' ' || e.last_name as manager_name
		FROM bank.branches b
		LEFT JOIN bank.employees e ON b.manager_id = e.employee_id
		WHERE b.branch_id = $1
	`
	var info models.BranchInfo
	err := s.db.QueryRowContext(ctx, query, branchID).Scan(
		&info.ID, &info.Name, &info.Location, &info.Phone, &info.Email, &info.ManagerName,
	)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (s *ReportService) getCustomerStats(ctx context.Context, branchID int64) (*models.CustomerStats, error) {
	query := `
		SELECT 
			COUNT(DISTINCT c.customer_id) as total_customers,
			COUNT(DISTINCT a.account_id) as total_accounts,
			COUNT(DISTINCT CASE WHEN a.status = 'ACTIVE' THEN a.account_id END) as active_accounts
		FROM bank.customers c
		LEFT JOIN bank.accounts a ON c.customer_id = a.customer_id
		WHERE c.branch_id = $1
	`
	var stats models.CustomerStats
	err := s.db.QueryRowContext(ctx, query, branchID).Scan(
		&stats.TotalCustomers, &stats.TotalAccounts, &stats.ActiveAccounts,
	)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (s *ReportService) getTransactionStats(ctx context.Context, branchID int64, month string) (*models.TransactionStats, error) {
	query := `
		SELECT 
			COALESCE(COUNT(*), 0) as total_transactions,
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(AVG(amount), 0) as average_amount
		FROM bank.transactions t
		JOIN bank.accounts a ON t.account_id = a.account_id
		JOIN bank.customers c ON a.customer_id = c.customer_id
		WHERE c.branch_id = $1
	`
	var stats models.TransactionStats
	err := s.db.QueryRowContext(ctx, query, branchID).Scan(
		&stats.TotalTransactions, &stats.TotalAmount, &stats.AverageAmount,
	)
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

func (s *ReportService) getDailyActivity(ctx context.Context, branchID int64, month string) ([]models.DailyActivity, error) {
	query := `
		WITH daily_stats AS (
			SELECT 
				DATE(t.created_at) as date,
				COALESCE(COUNT(*), 0) as transactions,
				COALESCE(SUM(amount), 0) as amount
			FROM bank.transactions t
			JOIN bank.accounts a ON t.account_id = a.account_id
			JOIN bank.customers c ON a.customer_id = c.customer_id
			WHERE c.branch_id = $1
			GROUP BY DATE(t.created_at)
		),
		prev_month_stats AS (
			SELECT 
				DATE(t.created_at) as date,
				COALESCE(COUNT(*), 0) as transactions,
				COALESCE(SUM(amount), 0) as amount
			FROM bank.transactions t
			JOIN bank.accounts a ON t.account_id = a.account_id
			JOIN bank.customers c ON a.customer_id = c.customer_id
			WHERE c.branch_id = $1
			GROUP BY DATE(t.created_at)
		)
		SELECT 
			ds.date,
			ds.transactions,
			ds.amount,
			CASE 
				WHEN pms.amount IS NULL OR pms.amount = 0 THEN 0
				ELSE ((ds.amount - pms.amount) / pms.amount) * 100
			END as growth_percent
		FROM daily_stats ds
		LEFT JOIN prev_month_stats pms ON ds.date = pms.date + INTERVAL '1 month'
		ORDER BY ds.date DESC
		LIMIT 30
	`
	rows, err := s.db.QueryContext(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []models.DailyActivity
	for rows.Next() {
		var activity models.DailyActivity
		err := rows.Scan(
			&activity.Date,
			&activity.Transactions,
			&activity.Amount,
			&activity.GrowthPercent,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, activity)
	}
	return activities, nil
}

func (s *ReportService) getTopCustomers(ctx context.Context, branchID int64, month string) ([]models.TopCustomer, error) {
	query := `
		SELECT 
			c.first_name || ' ' || c.last_name as name,
			COALESCE(COUNT(*), 0) as transactions,
			COALESCE(SUM(t.amount), 0) as total_amount
		FROM bank.transactions t
		JOIN bank.accounts a ON t.account_id = a.account_id
		JOIN bank.customers c ON a.customer_id = c.customer_id
		WHERE c.branch_id = $1
		GROUP BY c.customer_id, c.first_name, c.last_name
		ORDER BY total_amount DESC
		LIMIT 10
	`
	rows, err := s.db.QueryContext(ctx, query, branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var customers []models.TopCustomer
	for rows.Next() {
		var customer models.TopCustomer
		err := rows.Scan(
			&customer.Name,
			&customer.Transactions,
			&customer.TotalAmount,
		)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}
	return customers, nil
}
