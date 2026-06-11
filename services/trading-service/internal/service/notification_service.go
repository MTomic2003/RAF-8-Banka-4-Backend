package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/client"
	"github.com/RAF-SI-2025/Banka-4-Backend/services/trading-service/internal/model"
)

// notifyTimeout is the upper bound for a single notification round-trip
// (resolve recipient + send email). Notifications are best-effort and must not
// stall the caller, so this is intentionally tight.
const notifyTimeout = 10 * time.Second

// NotificationService produces and dispatches notification emails for order
// lifecycle events and price alerts. Every public method is fire-and-forget:
// the actual work runs in a goroutine with its own timeout, and any failure
// (recipient lookup, email transport, ...) is logged but never returned. This
// guarantees that a flaky email service or a missing user record cannot break
// the order or alert flow that triggered the notification.
type NotificationService struct {
	mailer     Mailer
	userClient client.UserServiceClient
}

func NewNotificationService(mailer Mailer, userClient client.UserServiceClient) *NotificationService {
	return &NotificationService{mailer: mailer, userClient: userClient}
}

// --- Order events ------------------------------------------------------------

func (n *NotificationService) NotifyOrderPending(order model.Order) {
	if n == nil {
		return
	}
	subject := fmt.Sprintf("Order #%d pending approval", order.OrderID)
	body := fmt.Sprintf(
		"Your %s order for %s (%d × %s) has been submitted and is awaiting supervisor approval.",
		order.Direction, orderTicker(order), order.Quantity, order.OrderType,
	)
	go n.sendForOrder(order, subject, body)
}

func (n *NotificationService) NotifyOrderApproved(order model.Order) {
	if n == nil {
		return
	}
	subject := fmt.Sprintf("Order #%d approved", order.OrderID)
	body := fmt.Sprintf(
		"Your %s order for %s (%d × %s) has been approved and will now be executed.",
		order.Direction, orderTicker(order), order.Quantity, order.OrderType,
	)
	go n.sendForOrder(order, subject, body)
}

func (n *NotificationService) NotifyOrderDeclined(order model.Order) {
	if n == nil {
		return
	}
	subject := fmt.Sprintf("Order #%d declined", order.OrderID)
	body := fmt.Sprintf(
		"Your %s order for %s (%d × %s) has been declined by the supervisor.",
		order.Direction, orderTicker(order), order.Quantity, order.OrderType,
	)
	go n.sendForOrder(order, subject, body)
}

func (n *NotificationService) NotifyOrderFilled(order model.Order) {
	if n == nil {
		return
	}
	pricePart := ""
	if order.PricePerUnit != nil {
		pricePart = fmt.Sprintf(" at %.2f per unit", *order.PricePerUnit)
	}
	subject := fmt.Sprintf("Order #%d fully executed", order.OrderID)
	body := fmt.Sprintf(
		"Your %s order for %s has been fully executed (%d units%s).",
		order.Direction, orderTicker(order), order.FilledQty, pricePart,
	)
	go n.sendForOrder(order, subject, body)
}

func (n *NotificationService) NotifyOrderPartialFill(order model.Order, fillQty uint, pricePerUnit float64) {
	if n == nil {
		return
	}
	subject := fmt.Sprintf("Order #%d partially filled", order.OrderID)
	body := fmt.Sprintf(
		"Your %s order for %s was partially filled (%d units at %.2f per unit). %d of %d units remaining.",
		order.Direction, orderTicker(order), fillQty, pricePerUnit,
		order.Quantity-order.FilledQty, order.Quantity,
	)
	go n.sendForOrder(order, subject, body)
}

func (n *NotificationService) NotifyOrderAutoCancelled(order model.Order) {
	if n == nil {
		return
	}
	subject := fmt.Sprintf("Order #%d cancelled", order.OrderID)
	body := fmt.Sprintf(
		"Your %s order for %s could not be executed and has been cancelled by the system.",
		order.Direction, orderTicker(order),
	)
	go n.sendForOrder(order, subject, body)
}

// --- Price alerts ------------------------------------------------------------

// NotifyPriceAlert delivers a price-alert email to the alert's owner. The
// listing argument carries the ticker for the email body; currentPrice is the
// observed price that triggered the alert.
func (n *NotificationService) NotifyPriceAlert(alert model.PriceAlert, listing *model.Listing, currentPrice float64) {
	if n == nil {
		return
	}
	ticker := "the asset"
	if listing != nil && listing.Asset != nil && listing.Asset.Ticker != "" {
		ticker = listing.Asset.Ticker
	}

	directionWord := "is now above"
	if alert.Condition == model.PriceAlertConditionBelow {
		directionWord = "is now below"
	}

	subject := fmt.Sprintf("Price alert: %s", ticker)
	body := fmt.Sprintf(
		"Your price alert has been triggered: %s %s your threshold of %.2f (current price: %.2f).",
		ticker, directionWord, alert.Threshold, currentPrice,
	)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), notifyTimeout)
		defer cancel()

		email, err := n.resolveOwnerEmail(ctx, alert.UserID, alert.OwnerType)
		if err != nil {
			log.Printf("[notify] price-alert %d: resolve email: %v", alert.PriceAlertID, err)
			return
		}
		if email == "" {
			log.Printf("[notify] price-alert %d: empty email for user %d (%s)", alert.PriceAlertID, alert.UserID, alert.OwnerType)
			return
		}
		if err := n.mailer.Send(email, subject, body); err != nil {
			log.Printf("[notify] price-alert %d: send email: %v", alert.PriceAlertID, err)
		}
	}()
}

// --- internals ---------------------------------------------------------------

func (n *NotificationService) sendForOrder(order model.Order, subject, body string) {
	ctx, cancel := context.WithTimeout(context.Background(), notifyTimeout)
	defer cancel()

	email, err := n.resolveOwnerEmail(ctx, order.OrderOwnerUserID, order.OrderOwnerType)
	if err != nil {
		log.Printf("[notify] order %d: resolve email: %v", order.OrderID, err)
		return
	}
	if email == "" {
		log.Printf("[notify] order %d: empty email for user %d (%s)", order.OrderID, order.OrderOwnerUserID, order.OrderOwnerType)
		return
	}
	if err := n.mailer.Send(email, subject, body); err != nil {
		log.Printf("[notify] order %d: send email: %v", order.OrderID, err)
	}
}

// resolveOwnerEmail looks up the user's email via UserServiceClient. Clients
// resolve through GetClientById, employees (actuaries / supervisors, who carry
// model.OwnerTypeBank since they act on behalf of the bank) through
// GetEmployeeById. Funds (model.OwnerTypeFund) are not real users with
// mailboxes, so they are skipped — for fund orders the employee placing the
// order is the OrderOwner and gets notified through the OwnerTypeBank path.
func (n *NotificationService) resolveOwnerEmail(ctx context.Context, userID uint, ownerType model.OwnerType) (string, error) {
	switch ownerType {
	case model.OwnerTypeClient:
		resp, err := n.userClient.GetClientById(ctx, uint64(userID))
		if err != nil {
			return "", err
		}
		return resp.GetEmail(), nil
	case model.OwnerTypeBank:
		resp, err := n.userClient.GetEmployeeById(ctx, uint64(userID))
		if err != nil {
			return "", err
		}
		return resp.GetEmail(), nil
	case model.OwnerTypeFund:
		// Funds do not have a mailbox; the actuary placing fund orders is
		// notified through their own employee record (OrderOwnerType=BANK).
		return "", nil
	default:
		return "", fmt.Errorf("unsupported owner type %q", ownerType)
	}
}

func orderTicker(order model.Order) string {
	if order.Listing.Asset != nil && order.Listing.Asset.Ticker != "" {
		return order.Listing.Asset.Ticker
	}
	return fmt.Sprintf("listing #%d", order.ListingID)
}
