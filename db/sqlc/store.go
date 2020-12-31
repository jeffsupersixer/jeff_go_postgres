package db

import (
	"context"
	"database/sql"
	"fmt"
)

// Store provides all functions to execute db queries and transactions
type Store struct {
	*Queries         // support single table query
	db       *sql.DB // we need this create transaction
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

func (store *Store) execTx(ctx context.Context, fn func(queries *Queries) error) error {
	tx, err := store.db.Begin()
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

type TransferTxParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResult struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

var txKey = struct {}{}

func (store *Store) TransferTx(ctx context.Context, arg TransferTxParams) (TransferTxResult, error) {
	var result TransferTxResult

	err := store.execTx(ctx, func(queries *Queries) error {
		var err error

		txName := ctx.Value(txKey)

		// create transfer between from and to account
		result.Transfer, err = queries.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: arg.FromAccountID,
			ToAccountID: arg.ToAccountID,
			Amount: arg.Amount,
		})
		fmt.Println("CreateTransfer: ", txName)
		if err != nil {
			return err
		}

		// create entry for from account
		result.FromEntry, err = queries.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.FromAccountID,
			Amount: -arg.Amount,
		})
		fmt.Println("CreateFromEntry: ", txName)
		if err != nil {
			return err
		}

		// create entry for to account
		result.ToEntry, err = queries.CreateEntry(ctx, CreateEntryParams{
			AccountID: arg.ToAccountID,
			Amount: arg.Amount,
		})
		fmt.Println("CreateToEntry: ", txName)
		if err != nil {
			return err
		}

		// get account -> update its balance
		/*account1, err := queries.GetAccountForUpdate(ctx, arg.FromAccountID)
		if err != nil {
			return err
		}
		result.FromAccount, err = queries.UpdateAccount(ctx, UpdateAccountParams{
			ID: account1.ID,
			Balance: account1.Balance - arg.Amount,
		})*/

		/*account2, err := queries.GetAccountForUpdate(ctx, arg.ToAccountID)
		if err != nil {
			return err
		}
		result.ToAccount, err = queries.UpdateAccount(ctx, UpdateAccountParams{
			ID: account2.ID,
			Balance: account2.Balance + arg.Amount,
		})*/

		// add balance directly
		// to avoid deadlock, do AddAccountBalance from lower ID to higher AccoutID

		if arg.FromAccountID < arg.ToAccountID {
			result.FromAccount, result.ToAccount, err = addMoney(ctx, queries, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount )
		} else {
			result.ToAccount, result.FromAccount, err = addMoney(ctx, queries, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount )
		}


		return nil
	})

	return result, err
}

func addMoney(ctx context.Context, q *Queries, accountID1, amount1, accountID2, amount2 int64) (account1, account2 Account, err error) {
	account1, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID: accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}

	account2, err = q.AddAccountBalance(ctx, AddAccountBalanceParams{
		ID: accountID2,
		Amount: amount2,
	})

	return
}
