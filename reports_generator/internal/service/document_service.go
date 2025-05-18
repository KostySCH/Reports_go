package service

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/KostySCH/Reports_go/reports_generator/internal/models"
	"github.com/jung-kurt/gofpdf"
	"github.com/nguyenthenguyen/docx"
)

type DocumentService struct {
	outputDir string
	fontsDir  string
}

func NewDocumentService(outputDir string) *DocumentService {
	workDir, err := os.Getwd()
	if err != nil {
		workDir = "."
	}
	projectRoot := filepath.Dir(workDir)
	return &DocumentService{
		outputDir: outputDir,
		fontsDir:  filepath.Join(projectRoot, "fonts", "dejavu-fonts-ttf-2.37", "ttf"),
	}
}

func (s *DocumentService) GenerateReport(data *models.BranchPerformanceData, format string) (string, error) {
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return "", fmt.Errorf("ошибка создания директории для отчетов: %v", err)
	}
	filename := fmt.Sprintf("branch_report_%d_%s.%s", data.BranchInfo.ID, time.Now().Format("20060102_150405"), format)
	filePath := filepath.Join(s.outputDir, filename)
	switch format {
	case "pdf":
		return s.generatePDF(data, filePath)
	case "docx":
		return s.generateDOCX(data, filePath)
	default:
		return "", fmt.Errorf("неподдерживаемый формат: %s", format)
	}
}

func (s *DocumentService) generatePDF(data *models.BranchPerformanceData, filePath string) (string, error) {
	regularFont := filepath.Join(s.fontsDir, "DejaVuSansCondensed.ttf")
	boldFont := filepath.Join(s.fontsDir, "DejaVuSansCondensed-Bold.ttf")
	if _, err := os.Stat(regularFont); os.IsNotExist(err) {
		return "", fmt.Errorf("шрифт не найден: %s", regularFont)
	}
	if _, err := os.Stat(boldFont); os.IsNotExist(err) {
		return "", fmt.Errorf("шрифт не найден: %s", boldFont)
	}
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddUTF8Font("DejaVu", "", regularFont)
	pdf.AddUTF8Font("DejaVu", "B", boldFont)
	pdf.SetFont("DejaVu", "", 12)
	pdf.AddPage()
	pdf.SetFont("DejaVu", "B", 16)
	pdf.Cell(190, 10, "Отчет по эффективности филиала")
	pdf.Ln(20)
	pdf.SetFont("DejaVu", "B", 14)
	pdf.Cell(190, 10, "Информация о филиале")
	pdf.Ln(10)
	pdf.SetFont("DejaVu", "", 12)
	pdf.Cell(190, 10, fmt.Sprintf("ID: %d", data.BranchInfo.ID))
	pdf.Ln(5)
	pdf.Cell(190, 10, fmt.Sprintf("Название: %s", data.BranchInfo.Name))
	pdf.Ln(5)
	pdf.Cell(190, 10, fmt.Sprintf("Адрес: %s", data.BranchInfo.Location))
	pdf.Ln(5)
	pdf.Cell(190, 10, fmt.Sprintf("Телефон: %s", data.BranchInfo.Phone))
	pdf.Ln(5)
	pdf.Cell(190, 10, fmt.Sprintf("Email: %s", data.BranchInfo.Email))
	pdf.Ln(5)
	pdf.Cell(190, 10, fmt.Sprintf("Менеджер: %s", data.BranchInfo.ManagerName))
	pdf.Ln(15)
	pdf.SetFont("DejaVu", "B", 14)
	pdf.Cell(190, 10, "Статистика клиентов")
	pdf.Ln(10)
	pdf.SetFont("DejaVu", "", 12)
	pdf.Cell(190, 10, fmt.Sprintf("Всего клиентов: %d", data.CustomerStats.TotalCustomers))
	pdf.Ln(5)
	pdf.Cell(190, 10, fmt.Sprintf("Всего счетов: %d", data.CustomerStats.TotalAccounts))
	pdf.Ln(5)
	pdf.Cell(190, 10, fmt.Sprintf("Активных счетов: %d", data.CustomerStats.ActiveAccounts))
	pdf.Ln(15)
	pdf.SetFont("DejaVu", "B", 14)
	pdf.Cell(190, 10, "Статистика транзакций")
	pdf.Ln(10)
	pdf.SetFont("DejaVu", "", 12)
	pdf.Cell(190, 10, fmt.Sprintf("Всего транзакций: %d", data.TransactionStats.TotalTransactions))
	pdf.Ln(5)
	pdf.Cell(190, 10, fmt.Sprintf("Общая сумма: %.2f ₽", data.TransactionStats.TotalAmount))
	pdf.Ln(5)
	pdf.Cell(190, 10, fmt.Sprintf("Средняя сумма: %.2f ₽", data.TransactionStats.AverageAmount))
	pdf.Ln(15)
	pdf.SetFont("DejaVu", "B", 14)
	pdf.Cell(190, 10, "Ежедневная активность")
	pdf.Ln(10)
	pdf.SetFont("DejaVu", "", 12)
	pdf.SetFillColor(240, 240, 240)
	pdf.Cell(50, 10, "Дата")
	pdf.Cell(40, 10, "Транзакции")
	pdf.Cell(50, 10, "Сумма")
	pdf.Cell(50, 10, "Рост")
	pdf.Ln(10)
	for _, activity := range data.DailyActivity {
		date, _ := time.Parse(time.RFC3339, activity.Date)
		pdf.Cell(50, 10, date.Format("02.01.2006"))
		pdf.Cell(40, 10, fmt.Sprintf("%d", activity.Transactions))
		pdf.Cell(50, 10, fmt.Sprintf("%.2f ₽", activity.Amount))
		pdf.Cell(50, 10, fmt.Sprintf("%.2f%%", activity.GrowthPercent))
		pdf.Ln(10)

	}
	if err := pdf.OutputFileAndClose(filePath); err != nil {
		return "", fmt.Errorf("ошибка сохранения PDF: %v", err)
	}
	return filePath, nil
}

func (s *DocumentService) generateDOCX(data *models.BranchPerformanceData, filePath string) (string, error) {
	// Получаем путь к шаблону
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("ошибка получения текущей директории: %v", err)
	}
	projectRoot := filepath.Dir(workDir)
	templatePath := filepath.Join(projectRoot, "templates", "reports_template.docx")

	// Создаем новый документ из шаблона
	r, err := docx.ReadDocxFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("ошибка чтения шаблона: %v", err)
	}
	defer r.Close()

	docx1 := r.Editable()

	// Информация о филиале
	docx1.Replace("{{branch_id}}", fmt.Sprintf("%d", data.BranchInfo.ID), -1)
	docx1.Replace("{{branch_name}}", data.BranchInfo.Name, -1)
	docx1.Replace("{{branch_location}}", data.BranchInfo.Location, -1)
	docx1.Replace("{{branch_phone}}", data.BranchInfo.Phone, -1)
	docx1.Replace("{{branch_email}}", data.BranchInfo.Email, -1)
	docx1.Replace("{{branch_manager}}", data.BranchInfo.ManagerName, -1)

	// Статистика клиентов
	docx1.Replace("{{total_customers}}", fmt.Sprintf("%d", data.CustomerStats.TotalCustomers), -1)
	docx1.Replace("{{total_accounts}}", fmt.Sprintf("%d", data.CustomerStats.TotalAccounts), -1)
	docx1.Replace("{{active_accounts}}", fmt.Sprintf("%d", data.CustomerStats.ActiveAccounts), -1)

	// Статистика транзакций
	docx1.Replace("{{total_transactions}}", fmt.Sprintf("%d", data.TransactionStats.TotalTransactions), -1)
	docx1.Replace("{{total_amount}}", fmt.Sprintf("%.2f ₽", data.TransactionStats.TotalAmount), -1)
	docx1.Replace("{{average_amount}}", fmt.Sprintf("%.2f ₽", data.TransactionStats.AverageAmount), -1)

	// Ежедневная активность
	var activityRows string
	for _, activity := range data.DailyActivity {
		date, _ := time.Parse(time.RFC3339, activity.Date)
		activityRows += fmt.Sprintf("<tr><td>%s</td><td>%d</td><td>%.2f ₽</td><td>%.2f%%</td></tr>",
			date.Format("02.01.2006"),
			activity.Transactions,
			activity.Amount,
			activity.GrowthPercent)
	}
	docx1.Replace("{{activity_rows}}", activityRows, -1)

	// Топ клиентов
	var customerRows string
	for _, customer := range data.TopCustomers {
		customerRows += fmt.Sprintf("<tr><td>%s</td><td>%d</td><td>%.2f ₽</td></tr>",
			customer.Name,
			customer.Transactions,
			customer.TotalAmount)
	}
	docx1.Replace("{{customer_rows}}", customerRows, -1)

	// Сохраняем документ
	if err := docx1.WriteToFile(filePath); err != nil {
		return "", fmt.Errorf("ошибка сохранения DOCX: %v", err)
	}

	return filePath, nil
}
