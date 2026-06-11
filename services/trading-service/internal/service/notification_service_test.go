package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/RAF-SI-2025/Banka-4-Backend/common/pkg/pb"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

// recordingMailer captures Send invocations so tests can assert on subject/body
// without spinning up a real email transport.
type recordingMailer struct {
	mu      sync.Mutex
	sent    []sentEmail
	failErr error
}

type sentEmail struct{ to, subject, body string }

func (r *recordingMailer) Send(to, subject, body string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.failErr != nil {
		return r.failErr
	}
	r.sent = append(r.sent, sentEmail{to: to, subject: subject, body: body})
	return nil
}

func (r *recordingMailer) snapshot() []sentEmail {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]sentEmail, len(r.sent))
	copy(out, r.sent)
	return out
}

type notifyUserClient struct {
	clientEmail   string
	clientErr     error
	employeeEmail string
	employeeErr   error
}

func (c *notifyUserClient) GetClientById(ctx context.Context, id uint64) (*pb.GetClientByIdResponse, error) {
	if c.clientErr != nil {
		return nil, c.clientErr
	}
	return &pb.GetClientByIdResponse{Id: id, Email: c.clientEmail}, nil
}

func (c *notifyUserClient) GetClientsByIds(ctx context.Context, ids []uint64) (*pb.GetClientsByIdsResponse, error) {
	return &pb.GetClientsByIdsResponse{}, nil
}

func (c *notifyUserClient) GetClientByIdentityId(ctx context.Context, identityId uint64) (*pb.GetClientByIdResponse, error) {
	return c.GetClientById(ctx, identityId)
}

func (c *notifyUserClient) GetEmployeeById(ctx context.Context, id uint64) (*pb.GetEmployeeByIdResponse, error) {
	if c.employeeErr != nil {
		return nil, c.employeeErr
	}
	return &pb.GetEmployeeByIdResponse{Id: id, Email: c.employeeEmail}, nil
}

func (c *notifyUserClient) GetEmployeeByIdentityId(ctx context.Context, identityId uint64) (*pb.GetEmployeeByIdResponse, error) {
	return c.GetEmployeeById(ctx, identityId)
}

func (c *notifyUserClient) GetAllClients(ctx context.Context, page, pageSize int32, firstName, lastName string) (*pb.GetAllClientsResponse, error) {
	return &pb.GetAllClientsResponse{}, nil
}

func (c *notifyUserClient) GetAllActuaries(ctx context.Context, page, pageSize int32, firstName, lastName string) (*pb.GetAllActuariesResponse, error) {
	return &pb.GetAllActuariesResponse{}, nil
}

func (c *notifyUserClient) GetIdentityByUserId(ctx context.Context, userID uint64, userType string) (*pb.GetIdentityByUserIdResponse, error) {
	return &pb.GetIdentityByUserIdResponse{}, nil
}

func (c *notifyUserClient) IncrementUsedLimit(ctx context.Context, employeeID uint64, amount float64) (*pb.IncrementUsedLimitResponse, error) {
	return &pb.IncrementUsedLimitResponse{}, nil
}

// waitForSends spins until the recording mailer has the expected count or the
// deadline elapses. NotifyOrder* fire goroutines, so the test needs to give
// them time to land before asserting.
func waitForSends(t *testing.T, mailer *recordingMailer, want int) []sentEmail {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got := mailer.snapshot()
		if len(got) >= want {
			return got
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected %d emails, got %d after timeout", want, len(mailer.snapshot()))
	return nil
}

func sampleOrder() model.Order {
	return model.Order{
		OrderID:          42,
		OrderOwnerUserID: 7,
		OrderOwnerType:   model.OwnerTypeClient,
		ListingID:        100,
		OrderType:        model.OrderTypeMarket,
		Direction:        model.OrderDirectionBuy,
		Quantity:         10,
		Listing: model.Listing{
			ListingID: 100,
			Asset:     &model.Asset{Ticker: "AAPL"},
		},
	}
}

func TestNotificationService_OrderPending_SendsToClient(t *testing.T) {
	t.Parallel()
	mailer := &recordingMailer{}
	users := &notifyUserClient{clientEmail: "alice@example.com"}
	n := NewNotificationService(mailer, users)

	n.NotifyOrderPending(sampleOrder())

	sent := waitForSends(t, mailer, 1)
	require.Len(t, sent, 1)
	assert.Equal(t, "alice@example.com", sent[0].to)
	assert.Contains(t, sent[0].subject, "pending approval")
	assert.Contains(t, sent[0].body, "AAPL")
}

func TestNotificationService_OrderApproved_UsesEmployeeLookupForBankOwner(t *testing.T) {
	t.Parallel()
	mailer := &recordingMailer{}
	users := &notifyUserClient{employeeEmail: "agent@example.com"}
	n := NewNotificationService(mailer, users)

	order := sampleOrder()
	order.OrderOwnerType = model.OwnerTypeBank
	n.NotifyOrderApproved(order)

	sent := waitForSends(t, mailer, 1)
	require.Len(t, sent, 1)
	assert.Equal(t, "agent@example.com", sent[0].to)
	assert.Contains(t, sent[0].subject, "approved")
}

func TestNotificationService_OrderPartialFill_IncludesQuantitiesAndPrice(t *testing.T) {
	t.Parallel()
	mailer := &recordingMailer{}
	users := &notifyUserClient{clientEmail: "alice@example.com"}
	n := NewNotificationService(mailer, users)

	order := sampleOrder()
	order.FilledQty = 4
	n.NotifyOrderPartialFill(order, 4, 150.5)

	sent := waitForSends(t, mailer, 1)
	require.Len(t, sent, 1)
	assert.Contains(t, sent[0].subject, "partially filled")
	assert.Contains(t, sent[0].body, "150.50")
	assert.Contains(t, sent[0].body, "4 units")
}

func TestNotificationService_FundOwnerType_NoEmailAttempted(t *testing.T) {
	t.Parallel()
	mailer := &recordingMailer{}
	users := &notifyUserClient{} // no emails configured; fund path skips lookup
	n := NewNotificationService(mailer, users)

	order := sampleOrder()
	order.OrderOwnerType = model.OwnerTypeFund
	n.NotifyOrderApproved(order)

	// Allow goroutine to finish; nothing should have been sent.
	time.Sleep(20 * time.Millisecond)
	assert.Empty(t, mailer.snapshot())
}

func TestNotificationService_NilReceiver_DoesNotPanic(t *testing.T) {
	t.Parallel()
	var n *NotificationService
	// All Notify* must be no-ops on a nil receiver — this is what lets unit
	// tests construct OrderService without a notifier.
	require.NotPanics(t, func() {
		n.NotifyOrderPending(sampleOrder())
		n.NotifyOrderApproved(sampleOrder())
		n.NotifyOrderDeclined(sampleOrder())
		n.NotifyOrderFilled(sampleOrder())
		n.NotifyOrderPartialFill(sampleOrder(), 1, 1.0)
		n.NotifyOrderAutoCancelled(sampleOrder())
		n.NotifyPriceAlert(model.PriceAlert{PriceAlertID: 1}, nil, 0)
	})
}

func TestNotificationService_PriceAlert_FormatsBodyWithTickerAndThreshold(t *testing.T) {
	t.Parallel()
	mailer := &recordingMailer{}
	users := &notifyUserClient{clientEmail: "alice@example.com"}
	n := NewNotificationService(mailer, users)

	alert := model.PriceAlert{
		PriceAlertID: 1,
		UserID:       7,
		OwnerType:    model.OwnerTypeClient,
		ListingID:    100,
		Condition:    model.PriceAlertConditionAbove,
		Threshold:    200,
	}
	listing := &model.Listing{Asset: &model.Asset{Ticker: "AAPL"}}
	n.NotifyPriceAlert(alert, listing, 205.0)

	sent := waitForSends(t, mailer, 1)
	require.Len(t, sent, 1)
	assert.Equal(t, "alice@example.com", sent[0].to)
	assert.Contains(t, sent[0].subject, "AAPL")
	assert.Contains(t, sent[0].body, "above")
	assert.Contains(t, sent[0].body, "200.00")
	assert.Contains(t, sent[0].body, "205.00")
}

func TestNotificationService_MailerError_DoesNotPanic(t *testing.T) {
	t.Parallel()
	mailer := &recordingMailer{failErr: errors.New("smtp down")}
	users := &notifyUserClient{clientEmail: "alice@example.com"}
	n := NewNotificationService(mailer, users)

	// Even when the mailer fails, the public API stays fire-and-forget — the
	// caller never sees the error and nothing panics.
	require.NotPanics(t, func() { n.NotifyOrderPending(sampleOrder()) })
	time.Sleep(20 * time.Millisecond)
}
