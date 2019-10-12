package main

import (
	"bufio"
	"flag"
	"github.com/mailru/dbr"
	_ "github.com/mailru/go-clickhouse"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var (
	dataPath = flag.String("data", "", "path to data file")
	chDsn    = flag.String("ch", "", "clickhouse dsn")
)

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

/**
 * статистика - ~400k в секунду
 * чтение укладывается в 0.1 сек:
SELECT *
FROM passports
WHERE code > 1000
LIMIT 100

Elapsed: 0.045 sec. Processed 65.54 thousand rows, 917.50 KB (1.45 million rows/s., 20.30 MB/s.)
*/
func main() {
	flag.Parse()

	connect, err := dbr.Open("clickhouse", "http://"+*chDsn+"/default", nil)
	handleErr(err)
	defer connect.Close()

	_, err = connect.Exec(`
		CREATE TABLE IF NOT EXISTS passports (
			code 		UInt32,
			num			UInt64,
			created_at	Date
		) ENGINE = MergeTree()
		ORDER BY (code, num)
	`)
	handleErr(err)

	tx, err := connect.Begin()
	handleErr(err)

	stmt, err := tx.Prepare(`INSERT INTO passports (code, num, created_at) VALUES (?, ?, ?)`)
	handleErr(err)

	fpath, err := filepath.Abs(*dataPath)
	handleErr(err)

	f, err := os.Open(fpath)
	handleErr(err)

	r := bufio.NewReader(f)
	_, _, err = r.ReadLine()
	handleErr(err)

	moment := time.Now()
	momentf := moment.Format(`2006-01-02`)
	var inx int

	go func() {
		log.Println("start parsing...")
		var last, diff, i int

		for {
			i++
			time.Sleep(5 * time.Second)
			diff = inx - last
			last = inx
			log.Printf("%04d. records: [%010d] +%d", i, inx, diff)
		}
	}()

	for {
		lineBytes, _, err := r.ReadLine()
		if err == io.EOF {
			break
		}
		handleErr(err)

		cols := strings.Split(*(*string)(unsafe.Pointer(&lineBytes)), `,`)
		code, err := strconv.Atoi(cols[0])
		if err != nil {
			continue
		}
		num, err := strconv.Atoi(cols[1])
		if err != nil {
			continue
		}

		_, err = stmt.Exec(code, num, momentf)
		handleErr(err)
		inx++
		if inx%500000 == 0 {
			err = tx.Commit()
			handleErr(err)

			tx, err = connect.Begin()
			handleErr(err)

			stmt, err = tx.Prepare(`INSERT INTO passports (code, num, created_at) VALUES (?, ?, ?)`)
			handleErr(err)
		}
	}
	err = tx.Commit()
	handleErr(err)

	log.Printf("\n\ndone! rec: [%d]\nduration: [%s]", inx, time.Since(moment))
}
