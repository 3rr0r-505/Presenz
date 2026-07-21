// main.go

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"

	"github.com/3rr0r-505/Presenz/internal/client"
	"github.com/3rr0r-505/Presenz/internal/config"
	"github.com/3rr0r-505/Presenz/internal/server"
	"github.com/3rr0r-505/Presenz/internal/services"
)

// stdinRouter owns the single os.Stdin reader for the whole process and
// routes each line to whichever channel is currently registered. Needed
// because two things want stdin at different times: the killswitch's
// "terminate" listener during normal operation, and the Ctrl+C y/N
// prompt when SIGINT fires. Python's asyncio sidesteps this because
// KeyboardInterrupt suspends the whole event loop; Go has no equivalent.
type stdinRouter struct {
	mu      sync.Mutex
	current chan<- string
}

func newStdinRouter() *stdinRouter {
	r := &stdinRouter{}
	go func() {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			r.mu.Lock()
			ch := r.current
			r.mu.Unlock()
			if ch != nil {
				select {
				case ch <- line:
				default:
				}
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "[ERROR] stdin read error:", err)
		}
	}()
	return r
}

func (r *stdinRouter) setTarget(ch chan<- string) {
	r.mu.Lock()
	r.current = ch
	r.mu.Unlock()
}

func main() {
	fmt.Println("\n[Presenz] Starting Presenz backend...")

	// -------------------------
	// Parse CLI arguments
	// -------------------------
	course := flag.String("course", "", "Course name")
	batch := flag.String("batch", "", "Batch ID")
	total := flag.Int("total", 0, "Total number of students")
	dbFlag := flag.String("db", "", "SQLite DB file path")
	exportFlag := flag.Bool("export", false, "Export attendance data on exit")
	flag.Parse()

	if *course == "" || *batch == "" || *total == 0 {
		fmt.Fprintln(os.Stderr, "[ERROR] --course, --batch, and --total are required")
		os.Exit(1)
	}

	const configPath = "config/config.json"
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[ERROR] Failed to load config")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	dbPath := *dbFlag
	if dbPath == "" {
		dbPath = cfg.DefaultDB()
	}
	fmt.Println("[Presenz] Using DB:", dbPath)

	// -------------------------
	// Initialize DB
	// -------------------------
	db, err := services.Connect(dbPath, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "[ERROR] Failed to connect to DB")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("[Presenz] DB connection established")

	// -------------------------
	// Initialize session
	// -------------------------
	session := &services.Session{}
	if err := session.Start(cfg, *total, *course, *batch, filepath.Base(dbPath)); err != nil {
		fmt.Fprintln(os.Stderr, "[ERROR] Failed to initialize session")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tableName := session.GetTableName()
	sessionCode := session.GetSessionCode()

	fmt.Println("+------------------------------------------------------------------------------------+")
	fmt.Println(" [Presenz] Session initialized:", tableName)
	fmt.Println(" [Presenz] Session code (share with students):", sessionCode)
	fmt.Println("+------------------------------------------------------------------------------------+")

	// -------------------------
	// Create attendance table
	// -------------------------
	if err := db.CreateTable(tableName); err != nil {
		fmt.Fprintln(os.Stderr, "[ERROR] Failed to create attendance table")
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("[Presenz] Attendance table created:", tableName)

	stdin := newStdinRouter()

	doExport := func() {
		records, err := db.FetchAll(tableName)
		if err != nil {
			fmt.Println("[ERROR] Export fetch failed:", err)
			return
		}
		if err := services.Export(cfg, tableName, records); err != nil {
			fmt.Println("[ERROR] Export failed:", err)
		}
	}

	fmt.Println("[Presenz] Presenz is ready to accept attendance submissions")

	// -------------------------
	// Outer run loop — mirrors Python's `while True`. Ctrl+C tears the
	// server down either way (asyncio.run does this in Python too, since
	// KeyboardInterrupt kills the running event loop); "resume" just
	// means rebuilding a fresh server + killswitch on the next iteration,
	// not literally pausing the same instance.
	// -------------------------
runLoop:
	for {
		ks := services.NewKillswitch(cfg.Killswitch.InactivityTimeoutMinutes)
		srv, err := server.New(cfg, session, db, ks, client.EntryHTML)
		if err != nil {
			fmt.Fprintln(os.Stderr, "[ERROR] Failed to start server")
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		termLines := make(chan string)
		stdin.setTarget(termLines)
		go ks.ManualTerminateListener(termLines)
		go ks.InactivityMonitor()

		serverErrCh := srv.Start()
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)

		select {
		case <-ks.WaitForShutdown():
			fmt.Println("[Presenz] KillSwitch triggered shutdown.")
			srv.Shutdown()
			signal.Stop(sigCh)
			if *exportFlag {
				doExport()
			}
			fmt.Println("[DEBUG] Server Halted gracefully.")
			break runLoop

		case err := <-serverErrCh:
			signal.Stop(sigCh)
			if err != nil {
				fmt.Println("[ERROR] Exception in server run:", err)
			}
			if *exportFlag {
				doExport()
			}
			fmt.Println("[DEBUG] Server Halted gracefully.")
			break runLoop

		case <-sigCh:
			signal.Stop(sigCh)
			fmt.Print("\n[Presenz] Ctrl+C detected. Do you want to quit? [y/N]: ")

			promptCh := make(chan string)
			stdin.setTarget(promptCh)
			answer := "n"
			if line, ok := <-promptCh; ok {
				answer = strings.ToLower(strings.TrimSpace(line))
			}

			srv.Shutdown()
			ks.TriggerShutdown() // stop this iteration's monitor/listener goroutines

			if answer == "y" {
				if *exportFlag {
					doExport()
				}
				db.Close()
				session.End()
				fmt.Println("[DEBUG] Server Halted gracefully.")
				fmt.Println("[Presenz] Shutting down...")
				break runLoop
			}

			fmt.Println("[Presenz] Resuming...")
			continue runLoop
		}
	}
}
