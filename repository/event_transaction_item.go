package repository

import (
	"assist-tix/config"
	"assist-tix/database"
	"assist-tix/model"
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type EventTransactionItemRepository interface {
	CreateTransactionItems(ctx context.Context, tx pgx.Tx, reqs []model.EventTransactionItem) (err error)
}

type EventTransactionItemRepositoryImpl struct {
	WrapDB *database.WrapDB
	Env    *config.EnvironmentVariable
}

func NewEventTransactionItemRepository(
	wrapDB *database.WrapDB,
	env *config.EnvironmentVariable,
) EventTransactionItemRepository {
	return &EventTransactionItemRepositoryImpl{
		WrapDB: wrapDB,
		Env:    env,
	}
}

func (r *EventTransactionItemRepositoryImpl) CreateTransactionItems(ctx context.Context, tx pgx.Tx, reqs []model.EventTransactionItem) (err error) {
	ctx, cancel := context.WithTimeout(ctx, r.Env.Database.Timeout.Write)
	defer cancel()

	query := `INSERT INTO event_transaction_items (
		transaction_id,
		garuda_id,

		quantity,

		full_name,
		email,
		phone_number,

		seat_row,
		seat_column,

		additional_information,
		total_price,
		
		created_at
	) VALUES `
	var args []interface{}
	var placeholders []string

	for i, req := range reqs {
		base := i * 10
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, NOW())",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10))

		args = append(args,
			req.TransactionID,
			req.GarudaID,
			req.Quantity,
			req.Fullname,
			req.Email,
			req.PhoneNumber,
			// req.TicketCategoryID,
			req.SeatRow,
			req.SeatColumn,
			req.AdditionalInformation,
			req.TotalPrice,
		)
	}

	query += strings.Join(placeholders, ",")

	if tx != nil {
		_, err = tx.Exec(ctx, query, args...)
	} else {
		_, err = r.WrapDB.Postgres.Exec(ctx, query, args...)
	}

	if err != nil {
		return
	}

	return
}
