# server/routes/__init__.py

"""
Routes package.

Contains all HTTP endpoint modules.
"""

from .attendance import router

__all__ = ["router"]