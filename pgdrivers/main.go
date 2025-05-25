package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
	"github.com/samber/lo"
)

const dsnTpl = "postgres://postgres:postgres@localhost:5454/postgres?sslmode=disable&%s"

var (
	mode        = flag.String("mode", "", "use mode: lib/pq | jackc/pgx | exec")
	showProblem = flag.Bool("problem", false, "")
	n           = flag.Int("n", 10, "")
)

func main() {
	flag.Parse()

	ctx, cnlFn := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cnlFn()

	switch *mode {
	case "lib/pq":
		runTestLibPq(ctx, !*showProblem)
	case "jackc/pgx":
		runTestPGXSimpleProtocol(ctx, lo.Ternary(*showProblem, "", "simple_protocol"))
	case "exec":
		runTestPGXSimpleProtocol(ctx, "exec")
	default:
		fmt.Printf("❌ unexpected mode value: %s\n", *mode)
	}
}

func runTest(ctx context.Context, db *sql.DB) {
	db.SetMaxOpenConns(2)

	wg := sync.WaitGroup{}
	wait := make(chan struct{})
	wg.Add(*n)
	for i := 0; i < *n; i++ {
		go func() {
			defer wg.Done()
			<-wait
			var (
				err  error
				rows *sql.Rows
			)

			argsN := rand.Intn(30) + 5
			params := lo.RepeatBy(argsN, func(index int) string {
				return fmt.Sprintf("$%d::int", index+1)
			})
			args := lo.RepeatBy(argsN, func(index int) any {
				return index
			})
			q := fmt.Sprintf("SELECT %s", strings.Join(params, "+"))
			rows, err = db.QueryContext(ctx, q, args...)
			if err != nil {
				fmt.Printf("❌ query failed: %s\n", err)
				return
			}
			println("✅ query executed successfully")
			rows.Close()
		}()
	}

	close(wait)
	wg.Wait()
}

func runTestLibPq(ctx context.Context, withBinaryParameters bool) {
	conn := lo.Must(
		pq.NewConnector(
			fmt.Sprintf(dsnTpl, lo.Ternary(withBinaryParameters, "binary_parameters=yes", "binary_parameters=no")),
		))

	db := sql.OpenDB(conn)
	defer db.Close()

	runTest(ctx, db)
}

func runTestPGXSimpleProtocol(ctx context.Context, execMode string) {
	if execMode != "" { // default value is cache_statement if execMode not set
		execMode = "default_query_exec_mode=" + execMode
	}

	dsn := fmt.Sprintf(dsnTpl, execMode)
	cfg := lo.Must(pgx.ParseConfig(dsn))
	db := stdlib.OpenDB(*cfg)
	defer db.Close()

	runTest(ctx, db)
}
