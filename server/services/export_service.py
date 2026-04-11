# server/services/export_service.py

import csv
import json
from pathlib import Path
from typing import List, Tuple

from server.config import settings

class ExportService:
    def export(self, table_name: str, records: List[Tuple]) -> None:
        backup_dir = Path(settings.backup_path)
        backup_dir.mkdir(parents=True, exist_ok=True)

        # CSV
        csv_path = backup_dir / f"{table_name}.csv"
        with csv_path.open("w", newline="", encoding="utf-8") as f:
            writer = csv.writer(f)
            writer.writerow(["name", "roll", "timestamp"])
            writer.writerows(records)

        # JSON
        json_path = backup_dir / f"{table_name}.json"
        structured = [
            {"name": r[0], "roll": r[1], "timestamp": r[2]}
            for r in records
        ]
        with json_path.open("w", encoding="utf-8") as f:
            json.dump(structured, f, indent=4)

        print(f"[Presenz] Exported to {csv_path}")
        print(f"[Presenz] Exported to {json_path}")


export_service = ExportService()