<div align="center">

# Presenz - FastAPI Powered Attendance System

![Python](https://img.shields.io/badge/Python-3.10+-417fb1?logo=python&logoColor=417fb1)
![FastAPI](https://img.shields.io/badge/FastAPI-Framework-019486?logo=fastapi&logoColor=019486)
![Uvicorn](https://img.shields.io/badge/Uvicorn-ASGI_Server-1ebdc8?logo=gunicorn&logoColor=1ebdc8)
![SQLite](https://img.shields.io/badge/Database-SQLite-3f9fdb?logo=sqlite&logoColor=3f9fdb)
![Docker](https://img.shields.io/badge/Docker-Supported-0091e2?logo=docker&logoColor=0091e2)
![Cloudflare Tunnel](https://img.shields.io/badge/Tunnel-Cloudflare-fbad41?logo=cloudflare&logoColor=fbad41)
![QRencode](https://img.shields.io/badge/QRencode-QR_Tool-000000?logo=matrix&logoColor=white)
![Interface](https://img.shields.io/badge/UI-Terminal-darkgreen?logo=gnubash)
![License](https://img.shields.io/badge/License-Apache%202.0-73e4bf?logo=opensourceinitiative&logoColor=73e4bf)

A lightweight, terminal-driven attendance system with secure public session access via Cloudflare Tunnel and QR-based student entry.

</div>

---

## 🔐 Key Features

- Minimal dependencies
- SQLite-based session storage with WAL mode
- Terminal-first workflow
- QR-driven attendance capture
- Docker-compatible deployment
- Ctrl+C protection — prompts confirmation before exit
- Per-IP rate limiting to prevent spam
- Built-in duplicate submission protection
- Atomic submission count — race-condition safe under concurrent load
- Export attendance as CSV and JSON on exit via `--export`
- Handles 200 concurrent requests with 99.5% success rate
- Maintains low average latency (~0.26–0.34s under load)
- Built-in duplicate submission protection (149/150 blocked in testing)

---

## 🛠️ Tech Stack

| Layer | Technology |
|---|---|
| Language | Python 3.10+ |
| Web Framework | FastAPI |
| ASGI Server | Uvicorn |
| Database | SQLite (WAL mode) |
| Input Validation | Pydantic |
| Rate Limiting | SlowAPI |
| Tunneling | Cloudflare Tunnel |
| QR Generation | qrencode (CLI) |
| Containerization | Docker |

## 📦 Requirements

- python3
- python3-venv
- sqlite3
- cloudflared
- qrencode
- Docker (optional)

---

## 🚀 Native Setup

### 1️⃣ Create Virtual Environment
```bash
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### 2️⃣ Start Presenz
```bash
python3 main.py --course <course-name> --batch <batch-name> --total <number-of-students> --db <path-to-db-file>
```

Example:

```bash
python3 main.py --course cs50 --batch fall-101 --total 120 --db db/ug-cs.db
```

### 3️⃣ Export Attendance on Exit (Optional)
Add `--export` to automatically save attendance as CSV and JSON to `backup/` when the server exits. Example:

```bash
python3 main.py --course cs50 --batch fall-101 --total 120 --db db/ug-cs.db --export
```

## 🐳 Docker Setup

### Build Image
```bash
docker build -t presenz-image:latest .
```

### Run Container
```bash
docker run -it --rm -p 8080:8080 -v $(pwd)/db:/app/db presenz-image:latest --course <course-name> --batch <batch-name> --total <number-of-students> --db <path-to-db-file>
```

Example:

```bash
docker run -it --rm -p 8080:8080 -v $(pwd)/db:/app/db presenz-image:latest --course cs50 --batch fall-101 --total 120 --db db/ug-cs.db
```

---

## 🌐 Public Access via Cloudflare Tunnel

Expose backend running on localhost:8080:

```bash
cloudflared tunnel --url http://localhost:8080
```

This will generate a public URL.

## 📲 Generate QR Code (Terminal)

```bash
qrencode -t ANSIUTF8 "<public-url>"
```

Students scan this QR to access the session.

---

## ⚙️ Configuration

All settings are managed via `config/config.json`:

| Key | Description | Default |
|---|---|---|
| `server.host` | Server bind address | `0.0.0.0` |
| `server.port` | Server port | `8080` |
| `database.wal_mode` | Enable SQLite WAL mode | `true` |
| `database.timeout_seconds` | SQLite connection timeout | `5` |
| `session.session_code_length` | Length of session code | `8` |
| `security.rate_limit` | Per-IP rate limit | `5/minute` |
| `killswitch.inactivity_timeout_minutes` | Auto-shutdown after inactivity | `3` |
| `export.backup_path` | Export output directory | `./backup/` |

## 🖥️ Server Controls

| Action | Method |
|---|---|
| Graceful exit | Type `terminate` in terminal |
| Confirmed exit | Press `Ctrl+C` → type `y` |
| Auto-shutdown | No activity for `inactivity_timeout_minutes` |

---

## 📄 License

This project is licensed under **Apache License 2.0**.
Free to use, modify, and learn from.

## ❤️ Support

If you like this project, consider giving it a ⭐ on GitHub!