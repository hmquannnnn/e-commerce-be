package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/hmquannnnn/e-commerce/payment-service/client"
	"github.com/hmquannnnn/e-commerce/payment-service/model"
	"github.com/hmquannnnn/e-commerce/payment-service/provider"
	"github.com/hmquannnnn/e-commerce/payment-service/repository"
)

type CreatePaymentInput struct {
	OrderID       string
	UserID        uuid.UUID
	Provider      model.Provider
	PaymentMethod model.PaymentMethod
	Amount        int64
	Currency      string
	ReturnURL     string
	CancelURL     string
}

type PaymentService interface {
	CreatePayment(ctx context.Context, input CreatePaymentInput) (*model.Payment, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Payment, error)
	GetByOrderID(ctx context.Context, orderID string, userID uuid.UUID) (*model.Payment, error)
	CancelPayment(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	HandlePayOSWebhook(req *http.Request) error
}

type paymentService struct {
	repo        repository.PaymentRepository
	providers   map[model.Provider]provider.PaymentProvider
	orderClient client.OrderClient
}

func NewPaymentService(repo repository.PaymentRepository, payosProvider provider.PaymentProvider, orderClient client.OrderClient) PaymentService {
	return &paymentService{
		repo: repo,
		providers: map[model.Provider]provider.PaymentProvider{
			model.ProviderPayOS: payosProvider,
		},
		orderClient: orderClient,
	}
}

const maxCreateCheckoutAttempts = 10

func (s *paymentService) CreatePayment(ctx context.Context, input CreatePaymentInput) (*model.Payment, error) {
	if input.OrderID == "" || input.Amount <= 0 {
		return nil, ErrInvalidInput
	}
	input.Currency = strings.ToUpper(input.Currency)
	if input.Currency == "" {
		input.Currency = "VND"
	}
	p, ok := s.providers[input.Provider]
	if !ok {
		return nil, ErrUnsupported
	}

	// ── Validate order với order-service ──────────────────────────────────────
	orderID, err := uuid.Parse(input.OrderID)
	if err != nil {
		return nil, ErrInvalidInput
	}
	orderInfo, err := s.orderClient.GetOrder(ctx, orderID)
	if err != nil {
		if errors.Is(err, client.ErrOrderNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("validate order: %w", err)
	}
	if orderInfo.UserID != input.UserID {
		return nil, ErrForbidden
	}
	if orderInfo.Status != "PENDING" {
		return nil, ErrOrderNotPayable
	}
	// So sánh amount với tolerance ±1 để tránh floating-point issue.
	// Order total_price là float64 (VND), payment amount là int64.
	expectedAmount := int64(math.Round(orderInfo.TotalPrice))
	if diff := input.Amount - expectedAmount; diff < -1 || diff > 1 {
		return nil, ErrAmountMismatch
	}
	// ─────────────────────────────────────────────────────────────────────────

	paymentID := uuid.New()

	if err := s.orderClient.TouchPaymentDeadline(ctx, orderID); err != nil {
		if errors.Is(err, client.ErrOrderNotPayable) {
			return nil, ErrOrderNotPayable
		}
		if errors.Is(err, client.ErrOrderNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("extend payment deadline: %w", err)
	}

	if existing, ok, err := s.getReusablePayment(ctx, input); err != nil {
		return nil, err
	} else if ok {
		return existing, nil
	}

	payosOrderCode, checkoutResp, err := s.createCheckoutWithRetry(ctx, p, paymentID, input)
	if err != nil {
		return nil, err
	}

	metadata, _ := json.Marshal(map[string]interface{}{
		"order_id": input.OrderID,
		"user_id":  input.UserID.String(),
	})
	payment := &model.Payment{
		ID:                     paymentID,
		OrderID:                input.OrderID,
		UserID:                 input.UserID,
		Provider:               input.Provider,
		PaymentMethod:          input.PaymentMethod,
		Amount:                 input.Amount,
		Currency:               input.Currency,
		Status:                 model.StatusPending,
		PayosOrderCode:         &payosOrderCode,
		ProviderPaymentID:      strPtr(checkoutResp.ProviderPaymentID),
		ProviderTransactionRef: strPtr(checkoutResp.ProviderTransactionRef),
		ReturnURL:              strPtr(input.ReturnURL),
		CancelURL:              strPtr(input.CancelURL),
		CheckoutURL:            strPtr(checkoutResp.CheckoutURL),
		ExpiresAt:              checkoutResp.ExpiresAt,
		Metadata:               metadata,
	}
	attempt := &model.PaymentAttempt{
		ID:                uuid.New(),
		PaymentID:         paymentID,
		AttemptNo:         1,
		ProviderSessionID: strPtr(checkoutResp.ProviderPaymentID),
		CheckoutURL:       strPtr(checkoutResp.CheckoutURL),
		Status:            "initiated",
		RawResponse:       checkoutResp.RawResponse,
	}
	if err := s.repo.CreatePaymentAndAttempt(ctx, payment, attempt); err != nil {
		return nil, fmt.Errorf("persist payment: %w", err)
	}
	return s.repo.GetByID(ctx, paymentID)
}

func (s *paymentService) getReusablePayment(ctx context.Context, input CreatePaymentInput) (*model.Payment, bool, error) {
	existing, err := s.repo.GetByOrderID(ctx, input.OrderID)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("check existing payment: %w", err)
	}

	if existing.UserID != input.UserID ||
		existing.Provider != input.Provider ||
		existing.PaymentMethod != input.PaymentMethod ||
		existing.Amount != input.Amount ||
		!strings.EqualFold(existing.Currency, input.Currency) ||
		existing.CheckoutURL == nil ||
		*existing.CheckoutURL == "" ||
		!isReusablePaymentStatus(existing.Status) {
		return nil, false, nil
	}

	return existing, true, nil
}

func isReusablePaymentStatus(status model.PaymentStatus) bool {
	return status == model.StatusPending ||
		status == model.StatusRequiresAction ||
		status == model.StatusProcessing
}

func (s *paymentService) createCheckoutWithRetry(
	ctx context.Context,
	p provider.PaymentProvider,
	paymentID uuid.UUID,
	input CreatePaymentInput,
) (int64, *provider.CheckoutResponse, error) {
	var lastErr error

	for attempt := 1; attempt <= maxCreateCheckoutAttempts; attempt++ {
		payosOrderCode, err := s.repo.NextPayosOrderCode(ctx)
		if err != nil {
			return 0, nil, fmt.Errorf("generate payos order code: %w", err)
		}

		checkoutResp, err := p.CreateCheckout(ctx, provider.CheckoutRequest{
			PaymentID:      paymentID,
			OrderID:        input.OrderID,
			UserID:         input.UserID,
			Amount:         input.Amount,
			Currency:       strings.ToLower(input.Currency),
			ReturnURL:      input.ReturnURL,
			CancelURL:      input.CancelURL,
			PaymentMethod:  input.PaymentMethod,
			Provider:       input.Provider,
			PayosOrderCode: payosOrderCode,
		})
		if err == nil {
			return payosOrderCode, checkoutResp, nil
		}
		if errors.Is(err, provider.ErrOrderCodeExists) {
			lastErr = err
			slog.Warn("payos order code already exists, retrying with a new code",
				"payos_order_code", payosOrderCode, "attempt", attempt)
			continue
		}
		return 0, nil, fmt.Errorf("create checkout: %w", err)
	}

	return 0, nil, fmt.Errorf("create checkout: exhausted payos order code retries: %w", lastErr)
}

func (s *paymentService) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*model.Payment, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	if p.UserID != userID {
		return nil, ErrPaymentNotFound
	}
	return p, nil
}

func (s *paymentService) GetByOrderID(ctx context.Context, orderID string, userID uuid.UUID) (*model.Payment, error) {
	p, err := s.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	if p.UserID != userID {
		return nil, ErrPaymentNotFound
	}
	return p, nil
}

func (s *paymentService) CancelPayment(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	if err := s.repo.CancelPending(ctx, id, userID); err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			return ErrPaymentNotFound
		}
		return err
	}
	return nil
}

func (s *paymentService) HandlePayOSWebhook(req *http.Request) error {
	cb, err := s.providers[model.ProviderPayOS].VerifyCallback(req)
	if err != nil {
		return err
	}

	// Resolve payment ID from the payos_order_code returned in webhook.
	payment, err := s.repo.GetByPayosOrderCode(req.Context(), cb.PayosOrderCode)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentNotFound) {
			// PayOS confirm-webhook hits this URL with signed sample payload; orderCode
			// is dummy and will not match any row. Return success so PayOS accepts the URL.
			slog.Info("payos webhook: unknown order code (confirm-webhook sample or stale)",
				"payos_order_code", cb.PayosOrderCode)
			return nil
		}
		return fmt.Errorf("resolve payment from payos order code %d: %w", cb.PayosOrderCode, err)
	}
	cb.PaymentID = payment.ID

	err = s.repo.ApplyCallback(req.Context(), model.ProviderPayOS, &repository.ApplyCallbackParams{
		EventID:                cb.EventID,
		PaymentID:              cb.PaymentID,
		Status:                 cb.Status,
		ProviderPaymentID:      cb.ProviderPaymentID,
		ProviderTransactionRef: cb.ProviderTransactionRef,
		RawPayload:             cb.RawPayload,
		SignatureValid:         cb.SignatureValid,
	})
	if errors.Is(err, repository.ErrDuplicateEvent) {
		return nil
	}
	if err != nil {
		return err
	}

	s.notifyOrderPaid(req.Context(), cb.PaymentID, cb.Status)
	return nil
}

// notifyOrderPaid asks order-service to flip PENDING → PAID. Order-service is
// the source of truth for the state machine, so this method only fires on
// success status — failures/expirations leave the order in PENDING for the
// expiry worker (or a retry from the user) to handle.
//
// Errors are logged but not propagated — the payment row is already persisted
// and the user will see the correct payment status.
func (s *paymentService) notifyOrderPaid(ctx context.Context, paymentID uuid.UUID, paymentStatus model.PaymentStatus) {
	if paymentStatus != model.StatusSucceeded {
		return
	}

	payment, err := s.repo.GetByID(ctx, paymentID)
	if err != nil {
		slog.Error("notify order paid: failed to load payment", "error", err, "payment_id", paymentID)
		return
	}

	orderID, err := uuid.Parse(payment.OrderID)
	if err != nil {
		slog.Error("notify order paid: invalid order_id on payment", "order_id", payment.OrderID, "payment_id", paymentID)
		return
	}

	if err := s.orderClient.MarkOrderPaid(ctx, orderID); err != nil {
		// ErrOrderNotPayable is informational, not a system failure: it means
		// the user (or admin) already moved the order off PENDING — most
		// often a cancellation race. We log so ops can investigate possible
		// money-taken-but-cancelled cases, but we don't return an error.
		slog.Error("notify order paid: failed to update order status",
			"error", err, "order_id", orderID, "payment_id", paymentID)
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
