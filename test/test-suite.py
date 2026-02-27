# robust_test_suite.py

import requests
from concurrent.futures import ThreadPoolExecutor, as_completed
import random
import string
import time
import threading
import csv

# -------------------------
# Configuration
# -------------------------
SERVER_URL = "http://localhost:8080/attendance/submit" # Replace with tunnel like if needed
SESSION_CODE = "<session code>"  # Replace with your current session code
TOTAL_STUDENTS = 200
MAX_WORKERS = 50
REQUEST_TIMEOUT = 5
MAX_DELAY = 3
MAX_RETRIES = 3

# Thread-safe print
print_lock = threading.Lock()

# -------------------------
# Helper Functions
# -------------------------
def random_name() -> str:
    first = ''.join(random.choices(string.ascii_letters, k=5))
    last = ''.join(random.choices(string.ascii_letters, k=5))
    return f"{first} {last}"

def random_roll(existing: set = None) -> str:
    """Generate unique roll number if set provided."""
    while True:
        part1 = ''.join(random.choices(string.ascii_uppercase + string.digits, k=3))
        part2 = ''.join(random.choices(string.digits, k=3))
        roll = f"{part1}-{part2}"
        if existing is None or roll not in existing:
            if existing is not None:
                existing.add(roll)
            return roll

def submit_request(payload) -> tuple[bool, float]:
    start = time.time()
    for attempt in range(1, MAX_RETRIES + 1):
        try:
            resp = requests.post(SERVER_URL, json=payload, timeout=REQUEST_TIMEOUT)
            elapsed = time.time() - start
            if resp.status_code == 200:
                return True, elapsed
            elif resp.status_code == 409:
                return False, elapsed
            else:
                time.sleep(0.1)
        except Exception:
            time.sleep(0.1)
    return False, REQUEST_TIMEOUT

# -------------------------
# Test Functions
# -------------------------
def stress_test():
    print(f"\nStarting Stress Test for {TOTAL_STUDENTS} students...")
    start_time = time.time()
    results = []
    roll_set = set()

    def task(i):
        time.sleep(random.uniform(0.1, MAX_DELAY))
        payload = {"name": random_name(), "roll": random_roll(roll_set), "session_code": SESSION_CODE}
        return submit_request(payload)

    with ThreadPoolExecutor(max_workers=MAX_WORKERS) as executor:
        futures = {executor.submit(task, i): i for i in range(1, TOTAL_STUDENTS + 1)}
        for future in as_completed(futures):
            i = futures[future]
            success, elapsed = future.result()
            with print_lock:
                print(f"[{i}] {'Success' if success else 'Failed'} - {elapsed:.2f}s")
            results.append((success, elapsed))

    # Summary
    total_time = time.time() - start_time
    successes = sum(1 for r in results if r[0])
    failures = TOTAL_STUDENTS - successes
    avg_latency = sum(r[1] for r in results) / TOTAL_STUDENTS

    print("\n+---------------- Stress Test Summary ----------------+")
    print(f"Total Requests   : {TOTAL_STUDENTS}")
    print(f"Successful      : {successes}")
    print(f"Failed          : {failures}")
    print(f"Average Latency : {avg_latency:.2f}s")
    print(f"Total Duration  : {total_time:.2f}s")
    print("+---------------------------------------------------+")

def latency_test():
    print("\nStarting Latency Test...")
    TOTAL = 200
    results = []

    def task(i):
        time.sleep(random.uniform(0.5, 5))
        payload = {"name": random_name(), "roll": random_roll(), "session_code": SESSION_CODE}
        return submit_request(payload)

    start_time = time.time()
    with ThreadPoolExecutor(max_workers=20) as executor:
        futures = {executor.submit(task, i): i for i in range(1, TOTAL + 1)}
        for future in as_completed(futures):
            i = futures[future]
            success, elapsed = future.result()
            with print_lock:
                print(f"[{i}] {'Success' if success else 'Failed'} - {elapsed:.2f}s")
            results.append((success, elapsed))

    total_time = time.time() - start_time
    successes = sum(1 for r in results if r[0])
    failures = TOTAL - successes
    avg_latency = sum(r[1] for r in results) / TOTAL

    print("\n+---------------- Latency Test Summary ----------------+")
    print(f"Total Requests   : {TOTAL}")
    print(f"Successful      : {successes}")
    print(f"Failed          : {failures}")
    print(f"Average Latency : {avg_latency:.2f}s")
    print(f"Total Duration  : {total_time:.2f}s")
    print("+---------------------------------------------------+")

def duplicate_test():
    print("\nStarting Duplicate Submission Test...")
    TOTAL = 150
    name = "John Doe"
    roll = "ABC-123"
    results = []

    def task(_):
        payload = {"name": name, "roll": roll, "session_code": SESSION_CODE}
        return submit_request(payload)

    start_time = time.time()
    with ThreadPoolExecutor(max_workers=20) as executor:
        futures = {executor.submit(task, i): i for i in range(TOTAL)}
        for future in as_completed(futures):
            success, elapsed = future.result()
            with print_lock:
                print(f"{'Success' if success else 'Failed'} - {elapsed:.2f}s")
            results.append((success, elapsed))

    successes = sum(1 for r in results if r[0])
    failures = TOTAL - successes
    total_time = time.time() - start_time

    print("\n+------------ Duplicate Submission Summary -----------+")
    print(f"Total Requests   : {TOTAL}")
    print(f"Accepted        : {successes}")
    print(f"Rejected/Failed : {failures}")
    print(f"Total Duration  : {total_time:.2f}s")
    print("+---------------------------------------------------+")

def invalid_session_test():
    print("\nStarting Invalid Session Code Test...")
    TOTAL = 150
    INVALID_CODES = ["XXXX", "1234ABCD", "INVALID", "ZZZZ9999"]
    results = []

    def task(_):
        payload = {"name": random_name(), "roll": random_roll(), "session_code": random.choice(INVALID_CODES)}
        return submit_request(payload)

    start_time = time.time()
    with ThreadPoolExecutor(max_workers=20) as executor:
        futures = {executor.submit(task, i): i for i in range(TOTAL)}
        for future in as_completed(futures):
            success, elapsed = future.result()
            with print_lock:
                print(f"{'Success' if success else 'Failed'} - {elapsed:.2f}s")
            results.append((success, elapsed))

    successes = sum(1 for r in results if r[0])
    failures = TOTAL - successes
    total_time = time.time() - start_time

    print("\n+------------ Invalid Session Summary --------------+")
    print(f"Total Requests   : {TOTAL}")
    print(f"Accepted        : {successes}")
    print(f"Rejected        : {failures}")
    print(f"Total Duration  : {total_time:.2f}s")
    print("+---------------------------------------------------+")

# -------------------------
# Menu
# -------------------------
def menu():
    while True:
        print("\n+----------------------------------------------------+")
        print("Select Test to Run:")
        print("1. Stress Test (Full student submission)")
        print("2. Latency Simulation")
        print("3. Duplicate Submission Test")
        print("4. Invalid Session Code Test")
        print("0. Exit")
        print("+----------------------------------------------------+\n")
        choice = input("Enter choice: ").strip()
        if choice == "1":
            stress_test()
        elif choice == "2":
            latency_test()
        elif choice == "3":
            duplicate_test()
        elif choice == "4":
            invalid_session_test()
        elif choice == "0":
            print("Exiting...")
            break
        else:
            print("Invalid choice, try again.")

if __name__ == "__main__":
    menu()