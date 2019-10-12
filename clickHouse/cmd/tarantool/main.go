package main

import (
	"bufio"
	"flag"
	"github.com/tarantool/go-tarantool"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

/**
 * статистика - ~20k в секунду
 */
const namespace = "passports"

var (
	dataPath = flag.String("data", "", "path to data file")
	tDsn     = flag.String("tarantool", "", "tarantool dsn")
	opts     = tarantool.Opts{User: "u", Pass: "tarPass"}
)

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type row struct {
	Code int
	Num  int
	Date string
}

func insertPool(cli *tarantool.Connection, ch <-chan row, poolSize int) {
	pool := make(chan struct{}, poolSize)
	for r := range ch {
		pool <- struct{}{}
		f := cli.InsertAsync(namespace, []interface{}{r.Code, r.Num, r.Date})
		go func(f *tarantool.Future, ch <-chan struct{}) {
			r, err := f.Get()
			if err != nil {
				log.Println(err)
			} else if r.Error != "" {
				log.Println(r.Error)
			}

			<-ch
		}(f, pool)
	}
	close(pool)
}

func main() {
	flag.Parse()
	/**
	  s = box.schema.space.create('passports')
	  s:format({{name = 'code', type = 'integer'},{name = 'num', type = 'integer'},{name = 'date', type = 'string'}})
	  s:create_index('primary', {type = 'hash',parts = {'code', 'num'}})

	*/
	cli, err := tarantool.Connect(*tDsn, opts)
	handleErr(err)

	defer cli.Close()

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
		var last, diff, i int

		for {
			i++
			time.Sleep(5 * time.Second)
			diff = inx - last
			last = inx
			log.Printf("%04d. records: [%010d] +%d", i, inx, diff)
		}
	}()

	pool := make(chan row, 100)
	go insertPool(cli, pool, 20)

	log.Println("start parsing...")

	var rec row
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

		rec.Code = code
		rec.Num = num
		rec.Date = momentf

		pool <- rec

		inx++
	}

	close(pool)
	log.Printf("\n\ndone! rec: [%d]\nduration: [%s]", inx, time.Since(moment))
}
