# server/config/settings.py

import json
import os
from pathlib import Path
from typing import Any, Dict


class Settings:
    def __init__(self) -> None:
        self._config = self._load_config()

    def _load_config(self) -> Dict[str, Any]:
        """
        Load configuration from config/config.json.
        Allows override via CONFIG_PATH environment variable.
        """
        config_path = os.getenv("CONFIG_PATH", "config/config.json")
        path = Path(config_path)

        if not path.exists():
            raise FileNotFoundError(f"Config file not found at {path}")

        with path.open("r", encoding="utf-8") as f:
            return json.load(f)

    # ------------------------
    # Server
    # ------------------------
    @property
    def server_host(self) -> str:
        return self._config["server"]["host"]

    @property
    def server_port(self) -> int:
        return self._config["server"]["port"]

    # ------------------------
    # Database
    # ------------------------
    @property
    def db_base_path(self) -> str:
        return self._config["database"]["base_path"]
    @property
    def default_db(self) -> str:
        return self.db_base_path + "default.db"

    @property
    def db_wal_mode(self) -> bool:
        return self._config["database"]["wal_mode"]

    @property
    def db_timeout(self) -> int:
        return self._config["database"]["timeout_seconds"]

    # ------------------------
    # Session
    # ------------------------
    @property
    def session_id_length(self) -> int:
        return self._config["session"]["session_id_length"]

    @property
    def session_code_length(self) -> int:
        return self._config["session"]["session_code_length"]

    # ------------------------
    # Security
    # ------------------------
    @property
    def max_name_length(self) -> int:
        return self._config["security"]["max_name_length"]

    @property
    def max_roll_length(self) -> int:
        return self._config["security"]["max_roll_length"]

    # ------------------------
    # Export
    # ------------------------
    @property
    def backup_path(self) -> str:
        return self._config["export"]["backup_path"]

    @property
    def export_format(self) -> str:
        return self._config["export"]["format"]


# Singleton instance
settings = Settings()