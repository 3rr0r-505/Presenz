# server/models/__init__.py

"""
Models package.

Contains:
- Pydantic schemas
- Data validation models
"""

from .schemas import AttendanceRequest

__all__ = ["AttendanceRequest"]