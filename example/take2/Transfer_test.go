package take2

import (
	"testing"
	"github.com/taowen/sqlxx"
	"github.com/stretchr/testify/require"
	"database/sql/driver"
	"github.com/go-sql-driver/mysql"
)

type subTest func(should *require.Assertions, conn *sqlxx.Conn)

var subTests map[string]subTest = map[string]subTest{
	"transfer success": func(should *require.Assertions, conn *sqlxx.Conn) {
		insert(should, conn, "account",
			"account_id", "acc1",
			"amount", int64(101))
		insert(should, conn, "account",
			"account_id", "acc2",
			"amount", int64(0))
		should.Nil(Transfer(conn, "acc1", "acc2", 100))
		should.Equal(100, queryAmount(should, conn, "acc2"))
		should.Equal(1, queryAmount(should, conn, "acc1"))
	},
	"not enough balance to transfer out": func(should *require.Assertions, conn *sqlxx.Conn) {
		insert(should, conn, "account",
			"account_id", "acc1",
			"amount", int64(0))
		insert(should, conn, "account",
			"account_id", "acc2",
			"amount", int64(0))
		should.NotNil(Transfer(conn, "acc1", "acc2", 100))
	},
}

func Test_transfer(t *testing.T) {
	for subTestName, subTest := range subTests {
		t.Run(subTestName, func(t *testing.T) {
			should := require.New(t)
			drv := mysql.MySQLDriver{}
			conn, err := sqlxx.Open(drv, "root:123456@tcp(127.0.0.1:3306)/take2")
			should.Nil(err)
			execute(should, conn, `
			CREATE TABLE IF NOT EXISTS account(
			account_id VARCHAR(128),
			amount INT,
			PRIMARY KEY (account_id)
			)`)
			execute(should, conn, `TRUNCATE TABLE account`)
			subTest(should, conn)
			conn.Close()
		})
	}
}


func insert(should *require.Assertions, conn *sqlxx.Conn, tableName string, inputs ... driver.Value) {
	columnNames := make([]interface{}, 0, len(inputs)/2)
	for i := 0; i < len(inputs); i += 2 {
		columnNames = append(columnNames, inputs[i])
	}
	stmt := conn.TranslateStatement(
		`INSERT INTO :STR_TABLE :INSERT_COLUMNS`, columnNames...)
	defer stmt.Close()
	inputs = append(inputs, "STR_TABLE")
	inputs = append(inputs, tableName)
	_, err := stmt.Exec(inputs...)
	should.Nil(err)
}

func execute(should *require.Assertions, conn *sqlxx.Conn, sql string) {
	stmt := conn.TranslateStatement(sql)
	defer stmt.Close()
	_, err := stmt.Exec()
	should.Nil(err)
}


func queryAmount(should *require.Assertions, conn *sqlxx.Conn, accountId string) int {
	stmt := conn.TranslateStatement(`SELECT * FROM account WHERE account_id=:account_id`)
	defer stmt.Close()
	rows, err := stmt.Query("account_id", accountId)
	should.Nil(err)
	defer rows.Close()
	should.Nil(rows.Next())
	return rows.GetInt(rows.C("amount"))
}