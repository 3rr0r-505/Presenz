// test/test-suite.go
//
// Go port of robust_test_suite.py — a live HTTP load-testing CLI for a
// running Presenz server. Not a unit test, not run via `go test`; this
// is a separate binary to run manually against a live instance.

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// -------------------------
// Configuration
// -------------------------
const (
	serverURL      = "http://localhost:8080/attendance/submit" // replace with tunnel url if needed
	sessionCode    = "5SYFAHCY"                                // replace with your current session code
	totalStudents  = 200
	maxWorkers     = 50
	requestTimeout = 5 * time.Second
	maxDelay       = 3 * time.Second
	maxRetries     = 3
)

var (
	printMu sync.Mutex
	client  = &http.Client{Timeout: requestTimeout}
)

type payload struct {
	Name        string `json:"name"`
	Roll        string `json:"roll"`
	SessionCode string `json:"session_code"`
}

// -------------------------
// Helper Functions
// -------------------------

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const upperAlnum = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const digits = "0123456789"

func randomString(charset string, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func randomName() string {
	first := randomString(letters, 5)
	last := randomString(letters, 5)
	return fmt.Sprintf("%s %s", first, last)
}

// rollGenerator produces unique rolls when a shared set + mutex are
// passed in (stress_test); nil set falls back to pure-random, unbounded
// collisions allowed, matching Python's random_roll() with no arg.
type rollGenerator struct {
	mu       sync.Mutex
	existing map[string]struct{}
}

func newRollGenerator() *rollGenerator {
	return &rollGenerator{existing: make(map[string]struct{})}
}

func (g *rollGenerator) next(unique bool) string {
	for {
		part1 := randomString(upperAlnum, 3)
		part2 := randomString(digits, 3)
		roll := fmt.Sprintf("%s-%s", part1, part2)

		if !unique {
			return roll
		}

		g.mu.Lock()
		if _, exists := g.existing[roll]; !exists {
			g.existing[roll] = struct{}{}
			g.mu.Unlock()
			return roll
		}
		g.mu.Unlock()
	}
}

// submitRequest posts a payload with retries, mirroring Python's
// submit_request: 200 -> success, 409 -> explicit failure (no retry
// needed, it's a definitive rejection), anything else -> retry with
// a short backoff, exhausting retries -> failure at timeout duration.
func submitRequest(p payload) (bool, time.Duration) {
	start := time.Now()

	body, _ := json.Marshal(p)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err := client.Post(serverURL, "application/json", bytes.NewReader(body))
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		resp.Body.Close()

		elapsed := time.Since(start)
		switch resp.StatusCode {
		case http.StatusOK:
			return true, elapsed
		case http.StatusConflict:
			return false, elapsed
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	return false, requestTimeout
}

type result struct {
	success bool
	elapsed time.Duration
}

// runConcurrent bounds concurrency to `workers` via a semaphore channel,
// mirroring ThreadPoolExecutor(max_workers=...), and runs `total` tasks.
// logFn is called under printMu for each completed task, in whatever
// order they finish (matches as_completed's non-deterministic order).
func runConcurrent(total, workers int, task func(i int) result, logFn func(i int, r result)) []result {
	results := make([]result, total)
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	for i := 0; i < total; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()

			r := task(i)
			results[i] = r

			printMu.Lock()
			logFn(i, r)
			printMu.Unlock()
		}(i)
	}

	wg.Wait()
	return results
}

func summarize(results []result) (successes, failures int, avgLatency, totalDuration time.Duration) {
	var sum time.Duration
	for _, r := range results {
		if r.success {
			successes++
		}
		sum += r.elapsed
	}
	failures = len(results) - successes
	if len(results) > 0 {
		avgLatency = sum / time.Duration(len(results))
	}
	return
}

// -------------------------
// Test Functions
// -------------------------

func stressTest() {
	fmt.Printf("\nStarting Stress Test for %d students...\n", totalStudents)
	start := time.Now()

	rolls := newRollGenerator()

	results := runConcurrent(totalStudents, maxWorkers, func(i int) result {
		time.Sleep(time.Duration(rand.Float64()*float64(maxDelay)) + 100*time.Millisecond)
		p := payload{Name: randomName(), Roll: rolls.next(true), SessionCode: sessionCode}
		success, elapsed := submitRequest(p)
		return result{success, elapsed}
	}, func(i int, r result) {
		status := "Failed"
		if r.success {
			status = "Success"
		}
		fmt.Printf("[%d] %s - %.2fs\n", i+1, status, r.elapsed.Seconds())
	})

	successes, failures, avgLatency, _ := summarize(results)
	totalTime := time.Since(start)

	fmt.Println("\n+---------------- Stress Test Summary ----------------+")
	fmt.Printf("Total Requests   : %d\n", totalStudents)
	fmt.Printf("Successful      : %d\n", successes)
	fmt.Printf("Failed          : %d\n", failures)
	fmt.Printf("Average Latency : %.2fs\n", avgLatency.Seconds())
	fmt.Printf("Total Duration  : %.2fs\n", totalTime.Seconds())
	fmt.Println("+---------------------------------------------------+")
}

func latencyTest() {
	const total = 200
	fmt.Println("\nStarting Latency Test...")
	start := time.Now()

	rolls := newRollGenerator()

	results := runConcurrent(total, 20, func(i int) result {
		delay := 500*time.Millisecond + time.Duration(rand.Float64()*4.5*float64(time.Second))
		time.Sleep(delay)
		p := payload{Name: randomName(), Roll: rolls.next(false), SessionCode: sessionCode}
		success, elapsed := submitRequest(p)
		return result{success, elapsed}
	}, func(i int, r result) {
		status := "Failed"
		if r.success {
			status = "Success"
		}
		fmt.Printf("[%d] %s - %.2fs\n", i+1, status, r.elapsed.Seconds())
	})

	successes, failures, avgLatency, _ := summarize(results)
	totalTime := time.Since(start)

	fmt.Println("\n+---------------- Latency Test Summary ----------------+")
	fmt.Printf("Total Requests   : %d\n", total)
	fmt.Printf("Successful      : %d\n", successes)
	fmt.Printf("Failed          : %d\n", failures)
	fmt.Printf("Average Latency : %.2fs\n", avgLatency.Seconds())
	fmt.Printf("Total Duration  : %.2fs\n", totalTime.Seconds())
	fmt.Println("+---------------------------------------------------+")
}

func duplicateTest() {
	const total = 150
	fmt.Println("\nStarting Duplicate Submission Test...")
	start := time.Now()

	name := "John Doe"
	roll := "ABC-123"

	results := runConcurrent(total, 20, func(i int) result {
		p := payload{Name: name, Roll: roll, SessionCode: sessionCode}
		success, elapsed := submitRequest(p)
		return result{success, elapsed}
	}, func(i int, r result) {
		status := "Failed"
		if r.success {
			status = "Success"
		}
		fmt.Printf("%s - %.2fs\n", status, r.elapsed.Seconds())
	})

	successes, failures, _, _ := summarize(results)
	totalTime := time.Since(start)

	fmt.Println("\n+------------ Duplicate Submission Summary -----------+")
	fmt.Printf("Total Requests   : %d\n", total)
	fmt.Printf("Accepted        : %d\n", successes)
	fmt.Printf("Rejected/Failed : %d\n", failures)
	fmt.Printf("Total Duration  : %.2fs\n", totalTime.Seconds())
	fmt.Println("+---------------------------------------------------+")
}

func invalidSessionTest() {
	const total = 150
	invalidCodes := []string{"XXXX", "1234ABCD", "INVALID", "ZZZZ9999"}
	fmt.Println("\nStarting Invalid Session Code Test...")
	start := time.Now()

	rolls := newRollGenerator()

	results := runConcurrent(total, 20, func(i int) result {
		p := payload{
			Name:        randomName(),
			Roll:        rolls.next(false),
			SessionCode: invalidCodes[rand.Intn(len(invalidCodes))],
		}
		success, elapsed := submitRequest(p)
		return result{success, elapsed}
	}, func(i int, r result) {
		status := "Failed"
		if r.success {
			status = "Success"
		}
		fmt.Printf("%s - %.2fs\n", status, r.elapsed.Seconds())
	})

	successes, failures, _, _ := summarize(results)
	totalTime := time.Since(start)

	fmt.Println("\n+------------ Invalid Session Summary --------------+")
	fmt.Printf("Total Requests   : %d\n", total)
	fmt.Printf("Accepted        : %d\n", successes)
	fmt.Printf("Rejected        : %d\n", failures)
	fmt.Printf("Total Duration  : %.2fs\n", totalTime.Seconds())
	fmt.Println("+---------------------------------------------------+")
}

// -------------------------
// Menu
// -------------------------

func menu() {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n+----------------------------------------------------+")
		fmt.Println("Select Test to Run:")
		fmt.Println("1. Stress Test (Full student submission)")
		fmt.Println("2. Latency Simulation")
		fmt.Println("3. Duplicate Submission Test")
		fmt.Println("4. Invalid Session Code Test")
		fmt.Println("0. Exit")
		fmt.Println("+----------------------------------------------------+")
		fmt.Print("\nEnter choice: ")

		if !scanner.Scan() {
			break
		}
		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			stressTest()
		case "2":
			latencyTest()
		case "3":
			duplicateTest()
		case "4":
			invalidSessionTest()
		case "all":
			stressTest()
			latencyTest()
			duplicateTest()
			invalidSessionTest()
		case "0":
			fmt.Println("Exiting...")
			return
		default:
			fmt.Println("Invalid choice, try again.")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "[ERROR] stdin read error:", err)
	}
}

func main() {
	// rand.Seed(time.Now().UnixNano())
	menu()
}
