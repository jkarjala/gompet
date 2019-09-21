package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jkarjala/gompet"

	_ "github.com/lib/pq"
)

// Long options for the testers, short ones used by the main library
var sqlDriver = flag.String("driver", "", "Database driver, 'postgres' or 'mysql'")
var sqlURL = flag.String("url", "", "SQL Connect URL, e.g. postgres://user:pass@host/db?sslmode=disable")
var sqlDiscard = flag.Bool("discard", false, "Discard result set with mimimal memory allocation")
var sqlTx = flag.Int("tx", 0, "Batch N commands in one transaction")

var discardResult = [][]string{{"discarded"}}

func main() {
	log.Println("SQL tester started")
	gompet.Run(clientFactory)
}

type myClient struct {
	id       int
	template *gompet.VarTemplate
	style    rune
	db       *sql.DB
	txCount  int
	tx       *sql.Tx
}

func clientFactory(id int, template string) (gompet.Client, error) {
	log.Println(id, "sql init", template)
	if *sqlDriver == "" || *sqlURL == "" {
		return nil, errors.New("Missing --driver and/or --url")
	}
	var style rune
	switch *sqlDriver {
	case "postgres":
		style = '$'
	case "mysql":
		style = '?'
	default:
		return nil, errors.New("Unsupported SQL driver")
	}

	db, err := sql.Open(*sqlDriver, *sqlURL)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	var client = myClient{id, gompet.Parse(template), style, db, 0, nil}
	return &client, nil
}

func (c *myClient) RunCommand(in *gompet.RunInput) *gompet.RunResult {
	query := in.Cmd
	var args []interface{}
	if c.template != nil {
		query, args = ExpandSQL(c.template, in.Args, '$')
	}

	var res string
	var err error
	var count int64
	var start = time.Now()
	if strings.ToLower(query[:6]) == "select" {
		if *sqlTx > 0 {
			return &gompet.RunResult{Err: errors.New("Transactions not supported for SELECT")}
		}
		var rows *sql.Rows
		rows, err = c.db.Query(query, args...)
		if err != nil {
			return &gompet.RunResult{Err: err}
		}
		rowResult, err := ReadRows(rows)
		if err != nil {
			return &gompet.RunResult{Err: err}
		}
		if *gompet.Verbose {
			log.Printf("%d sql select result: %s", c.id, rowResult)
		}
		count = int64(len(rowResult))
	} else {
		var result sql.Result
		if *sqlTx > 0 {
			if c.tx == nil {
				c.tx, err = c.db.Begin()
				if err != nil {
					return &gompet.RunResult{Err: err}
				}
			}
			result, err = c.tx.Exec(query, args...)
			if err != nil {
				c.tx.Rollback()
				return &gompet.RunResult{Err: err}
			}
			c.txCount++
			if c.txCount == *sqlTx {
				err = c.tx.Commit()
				if err != nil {
					return &gompet.RunResult{Err: err}
				}
				c.tx = nil
				c.txCount = 0
			}
		} else {
			result, err = c.db.Exec(query, args...)
			if err != nil {
				return &gompet.RunResult{Err: err}
			}
		}
		count, err = result.RowsAffected()
	}
	elapsed := time.Since(start).Seconds()
	res = fmt.Sprintf("%d rows", count)
	if *gompet.Verbose {
		log.Printf("%d sql %s %s: %v", c.id, query, args, res)
	}
	return &gompet.RunResult{Res: res, Time: elapsed}
}

func (c *myClient) Term() {
	log.Println(c.id, "sql term")
	if c.tx != nil { // This commit is not incldued in the final results...
		err := c.tx.Commit()
		if err != nil {
			log.Println("Final commit failed:", err)
		}
	}
	c.db.Close()
}

// ExpandSQL constructs SQL query string and arguments sorted to match it
func ExpandSQL(t *gompet.VarTemplate, args []string, style rune) (string, []interface{}) {
	var sqlArgs = make([]interface{}, 0)
	t.Builder.Reset()
	for i, piece := range t.Pieces {
		t.Builder.WriteString(piece)
		if i < len(t.Indices) && t.Indices[i]-1 < len(args) {
			switch style {
			case '$':
				t.Builder.WriteString("$")
				t.Builder.WriteByte(byte('0') + byte(i+1))
			case '?':
				t.Builder.WriteRune('?')
			default:
				panic(fmt.Sprintf("Unsupported style %c", style))
			}
			sqlArgs = append(sqlArgs, interface{}(args[t.Indices[i]-1]))
		}
	}
	return t.Builder.String(), sqlArgs
}

// ReadRows reads the resultset to an array of arrays of strings
func ReadRows(rows *sql.Rows) ([][]string, error) {
	var res [][]string
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	rawResult := make([][]byte, len(cols))
	dest := make([]interface{}, len(cols))
	for i := range rawResult {
		dest[i] = &rawResult[i]
	}

	count := 0
	for rows.Next() {
		err = rows.Scan(dest...)
		if err != nil {
			return nil, err
		}
		count++
		if *sqlDiscard {
			// discard result after reading
		} else {
			result := make([]string, len(cols))
			for i, raw := range rawResult {
				if raw == nil {
					result[i] = "null"
				} else {
					result[i] = string(raw)
				}
			}
			res = append(res, result)
		}
	}
	if *sqlDiscard {
		res = discardResult
	}
	return res, nil
}
