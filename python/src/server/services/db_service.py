# server/services/db_service.py

import sqlite3
from typing import List, Tuple

from server.config import settings


class DBService:
    def __init__(self) -> None:
        self._connection = None

    # ----------------------------
    # Connect
    # ----------------------------
    def connect(self, db_path: str) -> None:
        self._connection = sqlite3.connect(
            db_path,
            timeout=settings.db_timeout,
            check_same_thread=False,
        )

        if settings.db_wal_mode:
            self._connection.execute("PRAGMA journal_mode=WAL;")

    # ----------------------------
    # Table Creation
    # ----------------------------
    def create_table(self, table_name: str) -> None:
        query = f"""
        CREATE TABLE IF NOT EXISTS "{table_name}" (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            roll TEXT NOT NULL UNIQUE,
            timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        """
        self._connection.execute(query)
        self._connection.commit()

    # ----------------------------
    # Insert
    # ----------------------------
    def insert_attendance(self, table_name: str, name: str, roll: str) -> None:
        query = f'INSERT INTO "{table_name}" (name, roll) VALUES (?, ?);'
        self._connection.execute(query, (name, roll))
        self._connection.commit()

    # ----------------------------
    # Fetch
    # ----------------------------
    def fetch_all(self, table_name: str) -> List[Tuple]:
        cursor = self._connection.cursor()
        cursor.execute(f'SELECT name, roll, timestamp FROM "{table_name}";')
        return cursor.fetchall()

    # ----------------------------
    # Close
    # ----------------------------
    def close(self) -> None:
        if self._connection:
            self._connection.close()


db_service = DBService()