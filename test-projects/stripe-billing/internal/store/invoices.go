package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// InsertFromWebhook inserts an invoice received from a Stripe webhook.
// ON CONFLICT (stripe_invoice_id) DO NOTHING ensures idempotency — duplicate
// webhook deliveries are silently ignored.
func (s *Store) InsertFromWebhook(ctx context.Context, tx pgx.Tx, inv *Invoice) error {
	const q = `
		INSERT INTO invoices (
			workspace_id, stripe_invoice_id, amount_cents, currency,
			status, period_start, period_end, invoice_pdf_url
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (stripe_invoice_id) DO NOTHING`

	_, err := tx.Exec(ctx, q,
		inv.WorkspaceID,
		inv.StripeInvoiceID,
		inv.AmountPaid,
		inv.Currency,
		inv.Status,
		inv.PeriodStart,
		inv.PeriodEnd,
		inv.InvoicePDFURL,
	)
	if err != nil {
		return fmt.Errorf("store: insert invoice: %w", err)
	}
	return nil
}

// ListByWorkspace returns a page of invoices for the given workspace, sorted
// descending by created_at. It also returns the total row count for pagination.
func (s *Store) ListByWorkspace(ctx context.Context, workspaceID string, page, perPage int) ([]*Invoice, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	const countQ = `SELECT COUNT(*) FROM invoices WHERE workspace_id = $1`
	var total int
	if err := s.pool.QueryRow(ctx, countQ, workspaceID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("store: count invoices: %w", err)
	}

	const q = `
		SELECT id, workspace_id, stripe_invoice_id, amount_cents, currency,
		       status, period_start, period_end, invoice_pdf_url, created_at
		FROM invoices
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.pool.Query(ctx, q, workspaceID, perPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("store: list invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*Invoice
	for rows.Next() {
		inv := &Invoice{}
		if err := rows.Scan(
			&inv.ID,
			&inv.WorkspaceID,
			&inv.StripeInvoiceID,
			&inv.AmountPaid,
			&inv.Currency,
			&inv.Status,
			&inv.PeriodStart,
			&inv.PeriodEnd,
			&inv.InvoicePDFURL,
			&inv.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("store: scan invoice: %w", err)
		}
		invoices = append(invoices, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("store: iterate invoices: %w", err)
	}
	return invoices, total, nil
}
