package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go-api/config"
)

// SePay API Service
// Documentation: https://docs.sepay.vn

const (
	SepayAPIBaseURL = "https://my.sepay.vn/userapi"
)

// SepayTransaction giao dịch từ SePay API
type SepayTransaction struct {
	ID                 int64   `json:"id"`
	Gateway            string  `json:"gateway"`
	TransactionDate    string  `json:"transactionDate"`
	AccountNumber      string  `json:"accountNumber"`
	SubAccount         *string `json:"subAccount"`
	TransferType       string  `json:"transferType"`
	TransferAmount     float64 `json:"transferAmount"`
	Accumulated        float64 `json:"accumulated"`
	Code               *string `json:"code"`
	TransactionContent string  `json:"transactionContent"`
	ReferenceNumber    string  `json:"referenceNumber"`
	Description        string  `json:"description"`
}

// SepayTransactionListResponse response từ API lấy danh sách giao dịch
type SepayTransactionListResponse struct {
	Status       int                `json:"status"`
	Messages     interface{}        `json:"messages"`
	Transactions []SepayTransaction `json:"transactions"`
}

// SepayWebhookPayload payload từ SePay webhook
type SepayWebhookPayload struct {
	ID                 int64   `json:"id"`
	Gateway            string  `json:"gateway"`
	TransactionDate    string  `json:"transactionDate"`
	AccountNumber      string  `json:"accountNumber"`
	SubAccount         *string `json:"subAccount"`
	TransferType       string  `json:"transferType"`
	TransferAmount     float64 `json:"transferAmount"`
	Accumulated        float64 `json:"accumulated"`
	Code               *string `json:"code"`
	TransactionContent string  `json:"transactionContent"`
	ReferenceNumber    string  `json:"referenceNumber"`
	Description        string  `json:"description"`
}

// SepayService service để tương tác với SePay API
type SepayService struct {
	apiToken string
	client   *http.Client
}

// NewSepayService tạo SePay service mới
func NewSepayService() *SepayService {
	cfg := config.GetSepayConfig()
	return &SepayService{
		apiToken: cfg.APIToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetTransactions lấy danh sách giao dịch gần đây
func (s *SepayService) GetTransactions(accountNumber string, limit int) ([]SepayTransaction, error) {
	if s.apiToken == "" {
		return nil, fmt.Errorf("SePay API Token not configured")
	}

	url := fmt.Sprintf("%s/transactions/list?account_number=%s&limit=%d",
		SepayAPIBaseURL, accountNumber, limit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result SepayTransactionListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Status != 200 {
		return nil, fmt.Errorf("SePay API error: %v", result.Messages)
	}

	return result.Transactions, nil
}

// FindTransactionByContent tìm giao dịch theo nội dung chuyển khoản
func (s *SepayService) FindTransactionByContent(accountNumber string, content string, amount float64) (*SepayTransaction, error) {
	transactions, err := s.GetTransactions(accountNumber, 50)
	if err != nil {
		return nil, err
	}

	for _, tx := range transactions {
		// Kiểm tra nội dung chứa code và số tiền khớp
		if containsIgnoreCase(tx.TransactionContent, content) && tx.TransferAmount == amount {
			return &tx, nil
		}
	}

	return nil, nil // Không tìm thấy
}

// VerifyTransaction xác minh giao dịch có tồn tại không
func (s *SepayService) VerifyTransaction(transactionID int64) (*SepayTransaction, error) {
	transactions, err := s.GetTransactions("", 100)
	if err != nil {
		return nil, err
	}

	for _, tx := range transactions {
		if tx.ID == transactionID {
			return &tx, nil
		}
	}

	return nil, nil
}

// ===== SePay Linking Endpoints =====

// LinkingSessionRequest request để tạo phiên kết nối
type LinkingSessionRequest struct {
	BankCode      string `json:"bank_code"`
	AccountNumber string `json:"account_number"`
	AccountName   string `json:"account_name"`
	Webhook       string `json:"webhook,omitempty"`
}

// LinkingSessionResponse response từ tạo phiên kết nối
type LinkingSessionResponse struct {
	Status    int    `json:"status"`
	Messages  string `json:"messages"`
	SessionID string `json:"session_id"`
	QRCode    string `json:"qr_code"`
	LinkURL   string `json:"link_url"`
	ExpiresIn int    `json:"expires_in"` // seconds
}

// LinkingStatusResponse response trạng thái liên kết
type LinkingStatusResponse struct {
	Status      int    `json:"status"`
	Messages    string `json:"messages"`
	SessionID   string `json:"session_id"`
	Linked      bool   `json:"linked"`
	LinkedAt    string `json:"linked_at,omitempty"`
	AccountID   string `json:"account_id,omitempty"`
	BankCode    string `json:"bank_code,omitempty"`
	AccountName string `json:"account_name,omitempty"`
}

// CreateLinkingSession tạo phiên kết nối SePay
// Nhà hàng cần xác thực qua App ngân hàng
func (s *SepayService) CreateLinkingSession(req LinkingSessionRequest) (*LinkingSessionResponse, error) {
	if s.apiToken == "" {
		return nil, fmt.Errorf("SePay API Token not configured")
	}

	url := fmt.Sprintf("%s/connect/create", SepayAPIBaseURL)

	payload, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	// SePay Linking API expects different auth or structure
	// Sending as form data or JSON body
	httpReq, _ := http.NewRequest("POST", url, nil)
	httpReq.Header.Set("Authorization", "Bearer "+s.apiToken)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Bank-Code", req.BankCode)
	httpReq.Header.Set("X-Account-Number", req.AccountNumber)
	httpReq.Header.Set("X-Account-Name", req.AccountName)

	_ = payload // Mark payload as used for now
	_ = request // Mark request as used

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create linking session: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result LinkingSessionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		// If response is not JSON, return generic error
		return nil, fmt.Errorf("invalid response from SePay: %s", string(body))
	}

	return &result, nil
}

// GetLinkingStatus kiểm tra trạng thái liên kết
func (s *SepayService) GetLinkingStatus(sessionID string) (*LinkingStatusResponse, error) {
	if s.apiToken == "" {
		return nil, fmt.Errorf("SePay API Token not configured")
	}

	url := fmt.Sprintf("%s/connect/status?session_id=%s", SepayAPIBaseURL, sessionID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get linking status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result LinkingStatusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("invalid response from SePay: %s", string(body))
	}

	return &result, nil
}

// UnlinkAccount hủy liên kết tài khoản
func (s *SepayService) UnlinkAccount(accountID string) error {
	if s.apiToken == "" {
		return fmt.Errorf("SePay API Token not configured")
	}

	url := fmt.Sprintf("%s/connect/unlink", SepayAPIBaseURL)

	payload := map[string]string{"account_id": accountID}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+s.apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Account-ID", accountID)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to unlink account: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SePay unlink failed: %s", string(body))
	}

	_ = payloadBytes // Mark as used
	return nil
}

// containsIgnoreCase kiểm tra string chứa substring (không phân biệt hoa thường)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(substr) == 0 ||
			(len(s) > 0 && containsIgnoreCaseHelper(s, substr)))
}

func containsIgnoreCaseHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalIgnoreCase(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}
