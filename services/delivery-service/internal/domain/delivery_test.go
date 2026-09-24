package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDeliveryStatus_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status DeliveryStatus
		valid  bool
	}{
		{name: "pending is valid", status: DeliveryStatusPending, valid: true},
		{name: "in_transit is valid", status: DeliveryStatusInTransit, valid: true},
		{name: "ready_for_pickup is valid", status: DeliveryStatusReadyForPickup, valid: true},
		{name: "delivered is valid", status: DeliveryStatusDelivered, valid: true},
		{name: "cancelled is valid", status: DeliveryStatusCancelled, valid: true},
		{name: "invalid status", status: DeliveryStatus("INVALID"), valid: false},
		{name: "empty status", status: DeliveryStatus(""), valid: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.valid, tt.status.IsValid())
		})
	}
}

func TestDeliveryStatus_IsTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   DeliveryStatus
		terminal bool
	}{
		{name: "pending is not terminal", status: DeliveryStatusPending, terminal: false},
		{name: "in_transit is not terminal", status: DeliveryStatusInTransit, terminal: false},
		{name: "ready_for_pickup is not terminal", status: DeliveryStatusReadyForPickup, terminal: false},
		{name: "delivered is terminal", status: DeliveryStatusDelivered, terminal: true},
		{name: "cancelled is terminal", status: DeliveryStatusCancelled, terminal: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.terminal, tt.status.IsTerminal())
		})
	}
}

func TestCanTransition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		from    DeliveryStatus
		to      DeliveryStatus
		allowed bool
	}{
		{name: "pending -> in_transit", from: DeliveryStatusPending, to: DeliveryStatusInTransit, allowed: true},
		{name: "pending -> cancelled", from: DeliveryStatusPending, to: DeliveryStatusCancelled, allowed: true},
		{name: "pending -> delivered denied", from: DeliveryStatusPending, to: DeliveryStatusDelivered, allowed: false},
		{name: "pending -> ready_for_pickup denied", from: DeliveryStatusPending, to: DeliveryStatusReadyForPickup, allowed: false},
		{name: "pending -> pending denied", from: DeliveryStatusPending, to: DeliveryStatusPending, allowed: false},

		{name: "in_transit -> ready_for_pickup", from: DeliveryStatusInTransit, to: DeliveryStatusReadyForPickup, allowed: true},
		{name: "in_transit -> cancelled", from: DeliveryStatusInTransit, to: DeliveryStatusCancelled, allowed: true},
		{name: "in_transit -> delivered denied", from: DeliveryStatusInTransit, to: DeliveryStatusDelivered, allowed: false},
		{name: "in_transit -> pending denied", from: DeliveryStatusInTransit, to: DeliveryStatusPending, allowed: false},
		{name: "in_transit -> in_transit denied", from: DeliveryStatusInTransit, to: DeliveryStatusInTransit, allowed: false},

		{name: "ready_for_pickup -> delivered", from: DeliveryStatusReadyForPickup, to: DeliveryStatusDelivered, allowed: true},
		{name: "ready_for_pickup -> cancelled", from: DeliveryStatusReadyForPickup, to: DeliveryStatusCancelled, allowed: true},
		{name: "ready_for_pickup -> in_transit denied", from: DeliveryStatusReadyForPickup, to: DeliveryStatusInTransit, allowed: false},
		{name: "ready_for_pickup -> pending denied", from: DeliveryStatusReadyForPickup, to: DeliveryStatusPending, allowed: false},

		{name: "delivered -> any denied", from: DeliveryStatusDelivered, to: DeliveryStatusPending, allowed: false},
		{name: "delivered -> delivered denied", from: DeliveryStatusDelivered, to: DeliveryStatusDelivered, allowed: false},
		{name: "delivered -> cancelled denied", from: DeliveryStatusDelivered, to: DeliveryStatusCancelled, allowed: false},

		{name: "cancelled -> any denied", from: DeliveryStatusCancelled, to: DeliveryStatusPending, allowed: false},
		{name: "cancelled -> delivered denied", from: DeliveryStatusCancelled, to: DeliveryStatusDelivered, allowed: false},
		{name: "cancelled -> in_transit denied", from: DeliveryStatusCancelled, to: DeliveryStatusInTransit, allowed: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.allowed, CanTransition(tt.from, tt.to))
		})
	}
}

func TestCanReschedule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   DeliveryStatus
		expected bool
	}{
		{name: "pending allows reschedule", status: DeliveryStatusPending, expected: true},
		{name: "in_transit allows reschedule", status: DeliveryStatusInTransit, expected: true},
		{name: "ready_for_pickup denies reschedule", status: DeliveryStatusReadyForPickup, expected: false},
		{name: "delivered denies reschedule", status: DeliveryStatusDelivered, expected: false},
		{name: "cancelled denies reschedule", status: DeliveryStatusCancelled, expected: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, CanReschedule(tt.status))
		})
	}
}

func TestIsOverdue(t *testing.T) {
	t.Parallel()

	pastDate := time.Now().Add(-24 * time.Hour)
	futureDate := time.Now().Add(24 * time.Hour)
	now := time.Now()

	tests := []struct {
		name     string
		delivery *Delivery
		now      time.Time
		expected bool
	}{
		{
			name: "delivery on time - future estimated date",
			delivery: &Delivery{
				Status:                DeliveryStatusInTransit,
				EstimatedDeliveryDate: futureDate,
			},
			now:      now,
			expected: false,
		},
		{
			name: "delivery becomes overdue - past estimated date",
			delivery: &Delivery{
				Status:                DeliveryStatusInTransit,
				EstimatedDeliveryDate: pastDate,
			},
			now:      now,
			expected: true,
		},
		{
			name: "terminal status delivered - not overdue even with past date",
			delivery: &Delivery{
				Status:                DeliveryStatusDelivered,
				EstimatedDeliveryDate: pastDate,
			},
			now:      now,
			expected: false,
		},
		{
			name: "terminal status cancelled - not overdue even with past date",
			delivery: &Delivery{
				Status:                DeliveryStatusCancelled,
				EstimatedDeliveryDate: pastDate,
			},
			now:      now,
			expected: false,
		},
		{
			name: "pending status - overdue with past date",
			delivery: &Delivery{
				Status:                DeliveryStatusPending,
				EstimatedDeliveryDate: pastDate,
			},
			now:      now,
			expected: true,
		},
		{
			name: "rescheduled delivery - no longer overdue",
			delivery: &Delivery{
				Status:                DeliveryStatusInTransit,
				EstimatedDeliveryDate: futureDate,
			},
			now:      now,
			expected: false,
		},
		{
			name: "boundary - exactly at estimated date",
			delivery: &Delivery{
				Status:                DeliveryStatusPending,
				EstimatedDeliveryDate: now,
			},
			now:      now,
			expected: false,
		},
		{
			name: "boundary - one second after estimated date",
			delivery: &Delivery{
				Status:                DeliveryStatusPending,
				EstimatedDeliveryDate: now.Add(-1 * time.Second),
			},
			now:      now,
			expected: true,
		},
		{
			name: "ready_for_pickup - overdue with past date",
			delivery: &Delivery{
				Status:                DeliveryStatusReadyForPickup,
				EstimatedDeliveryDate: pastDate,
			},
			now:      now,
			expected: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.expected, IsOverdue(tt.delivery, tt.now))
		})
	}
}
