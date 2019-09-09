package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/jkarjala/gomb"

	_ "github.com/lib/pq"
)

// Long options for the testers, short ones used by the main library
var sqlDriver = flag.String("driver", "", "Database driver, 'postgres' or 'mysql'")
var sqlURL = flag.String("url", "", "SQL Connect URL")

func main() {
	log.Println("SQL tester started")
	gomb.Run(clientFactory)
}

type myClient struct {
	id       int
	template *gomb.VarTemplate
	style    rune
	db       *sql.DB
}

func clientFactory(id int, template string) (gomb.Client, error) {
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

	var client = myClient{id, gomb.Parse(template), style, db}
	return &client, nil
}

func (c *myClient) RunCommand(in gomb.RunInput) gomb.RunResult {
	query := in.Cmd
	var args []interface{}
	if c.template != nil {
		query, args = ExpandSQL(c.template, in.Args, '$')
	}

	var res string
	var err error
	var start = time.Now()
	result, err := c.db.Exec(query, args...)
	res = fmt.Sprintf("%s", result)
	elapsed := time.Since(start).Seconds()
	if err != nil {
		return gomb.RunResult{Err: err, Time: elapsed}
	}
	log.Printf("%d sql %s %s: %v", c.id, query, args, res)
	return gomb.RunResult{Res: res, Time: elapsed}
}

func (c *myClient) Term() {
	log.Println(c.id, "sql term")
}

// ExpandSQL constructs SQL query string and arguments sorted to match it
func ExpandSQL(t *gomb.VarTemplate, args []string, style rune) (string, []interface{}) {
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
