package main

import (
	"bufio"
	"flag"
	"gopkg.in/mgo.v2"
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
 * статистика - ~160k в секунду
 */
const part = 200000

var (
	dataPath = flag.String("data", "", "path to data file")
	dsn      = flag.String("mongo", "", "mongodb dsn")
)

type row struct {
	Code int
	Num  int
	Date time.Time
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()

	ses, err := mgo.Dial(*dsn)
	handleErr(err)
	db := ses.DB("main")
	c := db.C("passports")

	fpath, err := filepath.Abs(*dataPath)
	handleErr(err)

	f, err := os.Open(fpath)
	handleErr(err)

	r := bufio.NewReader(f)
	_, _, err = r.ReadLine()
	handleErr(err)

	moment := time.Now()
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

	var rec row
	bulk := c.Bulk()

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

		rec = row{Code: code, Num: num, Date: moment}
		bulk.Insert(rec)
		inx++

		if inx%part == 0 {
			_, err := bulk.Run()
			handleErr(err)
			bulk = c.Bulk()
		}
	}
	res, err := bulk.Run()
	handleErr(err)
	log.Println(res.Modified)

	log.Printf("\n\ndone! rec: [%d]\nduration: [%s]", inx, time.Since(moment))
}
