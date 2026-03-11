package notify

import "context"

// Notifier defines the interface for sending notification records, like Discord webhooks.
type Notifier interface {
	Notify(ctx context.Context, title string, message string, color int) error
}
