# server/services/session_service.py

import secrets
import string
from datetime import datetime
from typing import Optional

from server.config import settings


class SessionService:
    def __init__(self) -> None:
        self._active: bool = False
        self._session_id: Optional[str] = None
        self._session_code: Optional[str] = None
        self._table_name: Optional[str] = None
        self._max_count: int = 0
        self._current_count: int = 0
        self._db_path: Optional[str] = None

    # ----------------------------
    # Start Session
    # ----------------------------
    def start_session(
        self,
        max_count: int,
        course: str,
        batch: str,
        db_filename: str,
    ) -> None:
        if self._active:
            raise RuntimeError("Session already active")

        self._session_id = self._generate_token(settings.session_id_length)
        self._session_code = self._generate_token(settings.session_code_length)

        now = datetime.now().strftime("%d-%m-%y-%H%M")

        safe_course = self._sanitize_identifier(course)
        safe_batch = self._sanitize_identifier(batch)

        self._table_name = f"{now}-{safe_course}-{safe_batch}"

        self._max_count = max_count
        self._current_count = 0

        # Safe DB path resolution
        if "/" in db_filename or ".." in db_filename:
            raise ValueError("Invalid database filename")

        self._db_path = settings.db_base_path + db_filename
        self._active = True

    # ----------------------------
    # Utilities
    # ----------------------------
    def _generate_token(self, length: int) -> str:
        alphabet = string.ascii_uppercase + string.digits
        return "".join(secrets.choice(alphabet) for _ in range(length))

    def _sanitize_identifier(self, value: str) -> str:
        return "".join(c for c in value.upper() if c.isalnum())

    # ----------------------------
    # Validation
    # ----------------------------
    def validate_session_code(self, code: str) -> bool:
        return self._active and code == self._session_code

    def can_accept_submission(self) -> bool:
        return self._active and self._current_count < self._max_count

    def increment_count(self) -> None:
        self._current_count += 1

    def is_full(self) -> bool:
        return self._current_count >= self._max_count

    # ----------------------------
    # Getters
    # ----------------------------
    @property
    def get_session_code(self) -> str:
        return self._session_code

    @property
    def get_table_name(self) -> str:
        return self._table_name

    @property
    def db_path(self) -> str:
        return self._db_path

    @property
    def active(self) -> bool:
        return self._active

    # ----------------------------
    # End Session
    # ----------------------------
    def end_session(self) -> None:
        self._active = False
        self._session_id = None
        self._session_code = None
        self._table_name = None
        self._current_count = 0
        self._db_path = None


session_service = SessionService()