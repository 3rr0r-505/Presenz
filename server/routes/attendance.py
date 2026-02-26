# server/routes/attendance.py

import sqlite3
from fastapi import APIRouter, HTTPException # type: ignore
from fastapi.responses import FileResponse # type: ignore
from pathlib import Path

from server.models.schemas import AttendanceRequest
from server.security import (
    validate_name,
    validate_roll,
    validate_session_code,
    ValidationError,
)
from server.services.session_service import session_service
from server.services.db_service import db_service

router = APIRouter()

# -------------------------
# Serve entry.html
# -------------------------
BASE_DIR = Path(__file__).resolve().parent.parent.parent  # presenz root

@router.get("/")
def serve_entry():
    html_file = BASE_DIR / "client" / "entry.html"
    if not html_file.exists():
        raise HTTPException(status_code=404, detail="Entry form not found")
    return FileResponse(html_file)


# -------------------------
# Submit attendance
# -------------------------
@router.post("/submit")
def submit_attendance(payload: AttendanceRequest):
    try:
        # -------------------------
        # Validate Input
        # -------------------------
        name = validate_name(payload.name)
        roll = validate_roll(payload.roll)

        validate_session_code(
            payload.session_code,
            session_service.get_session_code,
        )

        table_name = session_service.get_table_name

        # -------------------------
        # Check submission limit
        # -------------------------
        if not session_service.can_accept_submission():
            return {
                "status": "closed",
                "message": "Attendance session closed",
            }

        # -------------------------
        # Insert Attendance
        # -------------------------
        db_service.insert_attendance(
            table_name=table_name,
            name=name,
            roll=roll,
        )
        session_service.increment_count()
        # If just reached max, print debug once
        if session_service.is_full():
            print("+------------------------------------------------------------+")
            print(f"| [DEBUG] All {session_service._max_count} responses submitted.                         |")
            print("+------------------------------------------------------------+")

        return {
            "status": "success",
            "message": "Attendance recorded successfully",
        }

    # Duplicate roll (UNIQUE constraint)
    except sqlite3.IntegrityError:
        raise HTTPException(
            status_code=409,
            detail="Roll number already submitted",
        )

    # Validation errors
    except ValidationError as e:
        raise HTTPException(
            status_code=400,
            detail=str(e),
        )

    # Any other unexpected error
    except Exception as e:
        print("[ERROR]", e)
        raise HTTPException(
            status_code=500,
            detail="Internal server error",
        )