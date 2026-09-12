package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"spinwheel/backend/pkg/models"
)

// PostgresWheelRepository persists wheels, spins, and events in Postgres
// (Neon). It implements WheelRepository. IDs are UUID strings; empty-string
// foreign keys (WheelID/SpinID on events) map to SQL NULL.
type PostgresWheelRepository struct {
	pool *pgxpool.Pool
}

var _ WheelRepository = (*PostgresWheelRepository)(nil)

func NewPostgresWheelRepository(ctx context.Context, databaseURL string) (*PostgresWheelRepository, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresWheelRepository{pool: pool}, nil
}

func (r *PostgresWheelRepository) Close() {
	r.pool.Close()
}

func (r *PostgresWheelRepository) CreateWheel(wheel *models.Wheel) (*models.Wheel, error) {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	wheelID := uuid.NewString()
	if _, err := tx.Exec(ctx,
		`INSERT INTO wheels (id, name) VALUES ($1::uuid, $2)`, wheelID, wheel.Name); err != nil {
		return nil, err
	}
	wheel.ID = wheelID

	for i := range wheel.Items {
		itemID := uuid.NewString()
		if _, err := tx.Exec(ctx,
			`INSERT INTO wheel_items (id, wheel_id, option, color, weight, position)
			 VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6)`,
			itemID, wheelID, wheel.Items[i].Option, wheel.Items[i].Color, wheel.Items[i].Weight, i); err != nil {
			return nil, err
		}
		wheel.Items[i].ID = itemID
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return wheel, nil
}

func (r *PostgresWheelRepository) GetWheel(id string) (*models.Wheel, error) {
	ctx := context.Background()
	wheelID, err := toPGUUID(id)
	if err != nil {
		return nil, err
	}

	var wheel models.Wheel
	var dbID pgtype.UUID
	if err := r.pool.QueryRow(ctx,
		`SELECT id, name FROM wheels WHERE id = $1`, wheelID).Scan(&dbID, &wheel.Name); err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	wheel.ID = fromPGUUID(dbID)

	items, err := r.listItems(ctx, wheelID)
	if err != nil {
		return nil, err
	}
	wheel.Items = items
	return &wheel, nil
}

func (r *PostgresWheelRepository) ListWheels() ([]*models.Wheel, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, `SELECT id, name FROM wheels ORDER BY created_at, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wheels []*models.Wheel
	for rows.Next() {
		var wheel models.Wheel
		var dbID pgtype.UUID
		if err := rows.Scan(&dbID, &wheel.Name); err != nil {
			return nil, err
		}
		wheel.ID = fromPGUUID(dbID)
		items, err := r.listItems(ctx, dbID)
		if err != nil {
			return nil, err
		}
		wheel.Items = items
		wheels = append(wheels, &wheel)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return wheels, nil
}

func (r *PostgresWheelRepository) UpdateWheel(wheel *models.Wheel) (*models.Wheel, error) {
	ctx := context.Background()
	wheelID, err := toPGUUID(wheel.ID)
	if err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	tag, err := tx.Exec(ctx, `UPDATE wheels SET name = $2 WHERE id = $1`, wheelID, wheel.Name)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}

	// Replace the item set (spins keep their winner_idx; winner_item_id
	// references go to NULL via ON DELETE SET NULL semantics on replace).
	if _, err := tx.Exec(ctx, `DELETE FROM wheel_items WHERE wheel_id = $1`, wheelID); err != nil {
		return nil, err
	}
	for i := range wheel.Items {
		itemID := wheel.Items[i].ID
		if _, err := uuid.Parse(itemID); err != nil {
			itemID = uuid.NewString()
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO wheel_items (id, wheel_id, option, color, weight, position)
			 VALUES ($1::uuid, $2, $3, $4, $5, $6)`,
			itemID, wheelID, wheel.Items[i].Option, wheel.Items[i].Color, wheel.Items[i].Weight, i); err != nil {
			return nil, err
		}
		wheel.Items[i].ID = itemID
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return wheel, nil
}

func (r *PostgresWheelRepository) DeleteWheel(id string) error {
	ctx := context.Background()
	wheelID, err := toPGUUID(id)
	if err != nil {
		return err
	}
	tag, err := r.pool.Exec(ctx, `DELETE FROM wheels WHERE id = $1`, wheelID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresWheelRepository) AddItem(wheelID string, item *models.WheelItem) (*models.WheelItem, error) {
	ctx := context.Background()
	wheelUUID, err := toPGUUID(wheelID)
	if err != nil {
		return nil, err
	}

	var exists bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM wheels WHERE id = $1)`, wheelUUID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}

	item.ID = uuid.NewString()
	if _, err := r.pool.Exec(ctx,
		`INSERT INTO wheel_items (id, wheel_id, option, color, weight, position)
		 VALUES ($1::uuid, $2, $3, $4, $5,
		         (SELECT COALESCE(MAX(position), -1) + 1 FROM wheel_items WHERE wheel_id = $2))`,
		item.ID, wheelUUID, item.Option, item.Color, item.Weight); err != nil {
		return nil, err
	}
	return item, nil
}

func (r *PostgresWheelRepository) UpdateItem(wheelID string, item *models.WheelItem) (*models.WheelItem, error) {
	ctx := context.Background()
	wheelUUID, err := toPGUUID(wheelID)
	if err != nil {
		return nil, err
	}
	itemUUID, err := toPGUUID(item.ID)
	if err != nil {
		return nil, err
	}

	tag, err := r.pool.Exec(ctx,
		`UPDATE wheel_items SET option = $3, color = $4, weight = $5
		 WHERE id = $1 AND wheel_id = $2`,
		itemUUID, wheelUUID, item.Option, item.Color, item.Weight)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrNotFound
	}
	return item, nil
}

func (r *PostgresWheelRepository) DeleteItem(wheelID string, itemID string) error {
	ctx := context.Background()
	wheelUUID, err := toPGUUID(wheelID)
	if err != nil {
		return err
	}
	itemUUID, err := toPGUUID(itemID)
	if err != nil {
		return err
	}

	tag, err := r.pool.Exec(ctx,
		`DELETE FROM wheel_items WHERE id = $1 AND wheel_id = $2`, itemUUID, wheelUUID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresWheelRepository) SpinWheel(id string) (*models.WheelItem, error) {
	wheel, err := r.GetWheel(id)
	if err != nil {
		return nil, err
	}
	idx, err := weightedWinnerIndex(wheel.Items)
	if err != nil {
		return nil, err
	}
	winningItem := wheel.Items[idx]
	return &winningItem, nil
}

func (r *PostgresWheelRepository) RecordSpin(ctx context.Context, params models.SpinParams) (*models.SpinRecord, error) {
	wheelUUID, err := toPGUUID(params.WheelID)
	if err != nil {
		return nil, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Serialize concurrent spins on the same wheel.
	var locked bool
	if err := tx.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM wheels WHERE id = $1 FOR UPDATE)`, wheelUUID).Scan(&locked); err != nil {
		return nil, err
	}
	if !locked {
		return nil, ErrNotFound
	}

	items, err := listItemsTx(ctx, tx, wheelUUID)
	if err != nil {
		return nil, err
	}

	idx, err := weightedWinnerIndex(items)
	if err != nil {
		return nil, err
	}

	seed, err := randomBytes(32)
	if err != nil {
		return nil, err
	}
	nonce, err := randomHex(16)
	if err != nil {
		return nil, err
	}

	expiresAt := params.ExpiresAt
	if expiresAt.IsZero() {
		expiresAt = time.Now().Add(60 * time.Second)
	}

	winnerItemUUID, err := toPGUUID(items[idx].ID)
	if err != nil {
		return nil, err
	}

	record := &models.SpinRecord{
		ID:           uuid.NewString(),
		WheelID:      params.WheelID,
		WinnerIdx:    idx,
		WinnerItemID: items[idx].ID,
		SeedServer:   seed,
		SeedClient:   params.SeedClient,
		ItemsHash:    params.ItemsHash,
		Nonce:        nonce,
		SessionID:    params.SessionID,
		IPHash:       params.IPHash,
		UA:           params.UA,
		CFRay:        params.CFRay,
		ExpiresAt:    expiresAt,
	}
	var dbID pgtype.UUID
	if err := tx.QueryRow(ctx,
		`INSERT INTO spins (id, wheel_id, winner_idx, winner_item_id, seed_server,
		                    seed_client, items_hash, nonce, session_id, ip_hash, ua, cf_ray, expires_at)
		 VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		 RETURNING id, created_at`,
		record.ID, wheelUUID, idx, winnerItemUUID, seed,
		params.SeedClient, params.ItemsHash, nonce,
		params.SessionID, params.IPHash, params.UA, params.CFRay, expiresAt,
	).Scan(&dbID, &record.CreatedAt); err != nil {
		return nil, err
	}
	record.ID = fromPGUUID(dbID)

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return record, nil
}

func (r *PostgresWheelRepository) UpdateSpinSig(ctx context.Context, spinID string, sig string) error {
	spinUUID, err := toPGUUID(spinID)
	if err != nil {
		return err
	}
	tag, err := r.pool.Exec(ctx, `UPDATE spins SET sig = $2 WHERE id = $1`, spinUUID, sig)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresWheelRepository) ListSpins(ctx context.Context, wheelID string, limit int) ([]*models.SpinRecord, error) {
	wheelUUID, err := toPGUUID(wheelID)
	if err != nil {
		return nil, err
	}

	var exists bool
	if err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM wheels WHERE id = $1)`, wheelUUID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrNotFound
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, wheel_id, winner_idx, winner_item_id, seed_server, seed_client,
		        items_hash, nonce, sig, session_id, ip_hash, ua, cf_ray, expires_at, created_at
		 FROM spins WHERE wheel_id = $1
		 ORDER BY created_at DESC, id DESC LIMIT $2`, wheelUUID, normalizeLimit(limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*models.SpinRecord
	for rows.Next() {
		record, err := scanSpin(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *PostgresWheelRepository) RecordEvent(ctx context.Context, params models.EventParams) (*models.EventRecord, error) {
	clientTS := params.ClientTS
	if clientTS.IsZero() {
		clientTS = time.Now()
	}
	wheelUUID := nullableUUID(params.WheelID)
	spinUUID := nullableUUID(params.SpinID)

	id := uuid.NewString()
	var (
		dbID      pgtype.UUID
		createdAt time.Time
	)
	err := r.pool.QueryRow(ctx,
		`INSERT INTO pageviews (id, session_id, type, wheel_id, spin_id, path, sig,
		                        client_ts, ip_hash, ua, cf_ray)
		 VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 ON CONFLICT DO NOTHING
		 RETURNING id, created_at`,
		id, params.SessionID, params.Type, wheelUUID, spinUUID, params.Path,
		params.Sig, clientTS, params.IPHash, params.UA, params.CFRay,
	).Scan(&dbID, &createdAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Idempotent redelivery: return the already-stored row.
			return r.getEvent(ctx, params.SessionID, params.Type, clientTS, params.SpinID)
		}
		return nil, err
	}

	return &models.EventRecord{
		ID:        fromPGUUID(dbID),
		SessionID: params.SessionID,
		Type:      params.Type,
		WheelID:   params.WheelID,
		SpinID:    params.SpinID,
		Path:      params.Path,
		Sig:       params.Sig,
		ClientTS:  clientTS,
		IPHash:    params.IPHash,
		UA:        params.UA,
		CFRay:     params.CFRay,
		CreatedAt: createdAt,
	}, nil
}

func (r *PostgresWheelRepository) getEvent(ctx context.Context, sessionID, eventType string, clientTS time.Time, spinID string) (*models.EventRecord, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, session_id, type, wheel_id, spin_id, path, sig,
		        client_ts, ip_hash, ua, cf_ray, created_at
		 FROM pageviews
		 WHERE session_id = $1 AND type = $2 AND client_ts = $3
		   AND COALESCE(spin_id::text, '') = $4`,
		sessionID, eventType, clientTS, spinID)
	record, err := scanEvent(row)
	if err != nil {
		return nil, err
	}
	return record, nil
}

func (r *PostgresWheelRepository) listItems(ctx context.Context, wheelUUID pgtype.UUID) ([]models.WheelItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, option, color, weight FROM wheel_items
		 WHERE wheel_id = $1 ORDER BY position, id`, wheelUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectItems(rows)
}

func listItemsTx(ctx context.Context, tx pgx.Tx, wheelUUID pgtype.UUID) ([]models.WheelItem, error) {
	rows, err := tx.Query(ctx,
		`SELECT id, option, color, weight FROM wheel_items
		 WHERE wheel_id = $1 ORDER BY position, id`, wheelUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectItems(rows)
}

func collectItems(rows pgx.Rows) ([]models.WheelItem, error) {
	items := []models.WheelItem{}
	for rows.Next() {
		var (
			item models.WheelItem
			id   pgtype.UUID
		)
		if err := rows.Scan(&id, &item.Option, &item.Color, &item.Weight); err != nil {
			return nil, err
		}
		item.ID = fromPGUUID(id)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type scannableRow interface {
	Scan(dest ...any) error
}

func scanSpin(row scannableRow) (*models.SpinRecord, error) {
	var (
		record       models.SpinRecord
		dbID         pgtype.UUID
		dbWheelID    pgtype.UUID
		dbWinnerItem pgtype.UUID
		expiresAt    pgtype.Timestamptz
	)
	if err := row.Scan(&dbID, &dbWheelID, &record.WinnerIdx, &dbWinnerItem,
		&record.SeedServer, &record.SeedClient, &record.ItemsHash, &record.Nonce,
		&record.Sig, &record.SessionID, &record.IPHash, &record.UA, &record.CFRay,
		&expiresAt, &record.CreatedAt); err != nil {
		return nil, err
	}
	record.ID = fromPGUUID(dbID)
	record.WheelID = fromPGUUID(dbWheelID)
	record.WinnerItemID = fromPGUUID(dbWinnerItem)
	if expiresAt.Valid {
		record.ExpiresAt = expiresAt.Time
	}
	return &record, nil
}

func scanEvent(row scannableRow) (*models.EventRecord, error) {
	var (
		record    models.EventRecord
		dbID      pgtype.UUID
		dbWheelID pgtype.UUID
		dbSpinID  pgtype.UUID
	)
	if err := row.Scan(&dbID, &record.SessionID, &record.Type, &dbWheelID, &dbSpinID,
		&record.Path, &record.Sig, &record.ClientTS,
		&record.IPHash, &record.UA, &record.CFRay, &record.CreatedAt); err != nil {
		return nil, err
	}
	record.ID = fromPGUUID(dbID)
	record.WheelID = fromPGUUID(dbWheelID)
	record.SpinID = fromPGUUID(dbSpinID)
	return &record, nil
}

func toPGUUID(s string) (pgtype.UUID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: [16]byte(u), Valid: true}, nil
}

// nullableUUID maps "" to SQL NULL so optional event keys stay unset.
func nullableUUID(s string) pgtype.UUID {
	if s == "" {
		return pgtype.UUID{Valid: false}
	}
	if u, err := uuid.Parse(s); err == nil {
		return pgtype.UUID{Bytes: [16]byte(u), Valid: true}
	}
	return pgtype.UUID{Valid: false}
}

func fromPGUUID(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	id, err := uuid.FromBytes(u.Bytes[:])
	if err != nil {
		return ""
	}
	return id.String()
}
