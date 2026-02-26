# server/services/export_service.py

import json
from pathlib import Path
from typing import List, Tuple

from server.config import settings


class ExportService:
    def export_json(
        self,
        table_name: str,
        records: List[Tuple],
    ) -> str:
        backup_dir = Path(settings.backup_path)
        backup_dir.mkdir(parents=True, exist_ok=True)

        export_path = backup_dir / f"{table_name}.json"

        structured = [
            {
                "name": r[0],
                "roll": r[1],
                "timestamp": r[2],
            }
            for r in records
        ]

        with export_path.open("w", encoding="utf-8") as f:
            json.dump(structured, f, indent=4)

        return str(export_path)


export_service = ExportService()