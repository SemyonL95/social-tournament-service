package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func queryBuildWhereInSelectUsers(IDs []string) (string, []interface{}) {
	queryStr := ""
	var tempBackersIDs []interface{}

	for i, _ := range IDs {
		queryStr += fmt.Sprint("$", i+1, ",")
		tempBackersIDs = append(tempBackersIDs, IDs[i])
	}
	queryStr = queryStr[0: len(queryStr)-1]
	sqlSelectUsers := fmt.Sprintf("SELECT * FROM users WHERE id IN (%s) FOR UPDATE", queryStr)

	return sqlSelectUsers, tempBackersIDs
}

func transact(db *sqlx.DB, txFunc func(*sqlx.Tx) error) (err error) {
	tx, err := db.Beginx()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
			if err != nil {
				err = tx.Rollback()
			}
		}
	}()
	err = txFunc(tx)
	return err
}

//We retry on those two error codes. This is due to serializable isolation level transactions set
func retryTransact(db *sqlx.DB, txFunc func(*sqlx.Tx) error, retryNumber int) error {
	var err error
	for i := 0; i < retryNumber; i++ {
		err = transact(db, txFunc)
		if err == nil {
			return err
		}
		driverError, ok := err.(*pq.Error)
		if !ok {
			return err
		}

		//40001 serialization_failure
		//40P01 deadlock_detected
		if driverError.Code != "40001" && driverError.Code != "40P01" {
			return err
		}
	}
	return err
}
