# Presenzc - FastAPI Powered Attendance System

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

## 🔐 Key Features

- Minimal dependencies  
- SQLite-based session storage  
- Terminal-first workflow  
- QR-driven attendance capture  
- Docker-compatible deployment  
- Handles 200 concurrent requests with 99.5% success rate  
- Maintains low average latency (~0.26–0.34s under load)  
- Built-in duplicate submission protection (149/150 blocked in testing)  

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

`python3 main.py --course cs50 --batch fall-101 --total 120 --db db/ug-cs.db`

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

`docker run -it --rm -p 8080:8080 -v $(pwd)/db:/app/db presenz-image:latest --course cs50 --batch fall-101 --total 120 --db db/ug-cs.db`

---

## 🌐 Public Access via Cloudflare Tunnel

Expose backend running on localhost:8080:
```bash
cloudflared --tunnel http://localhost:8080
```
This will generate a public URL.

## 📲 Generate QR Code (Terminal)
```bash
qrencode -t ANSIUTF8 "<public-url>"
```

Students scan this QR to access the session.

## 💾 Backup Session Data (SQLite → JSON)

Export a session table:
```bash
sqlite3 -json "<path-to-db-file>" 'SELECT * FROM "<table-name>";' > backup/"table-name".json
```

## ⚙️ Default Server
The application runs on Uvicorn ASGI at `http://0.0.0.0:8080`

---

## 📄 License
This project is licensed under **Apache License 2.0**.
Free to use, modify, and learn from.

## ❤️ Support
If you like this project, consider giving it a ⭐ on GitHub!
