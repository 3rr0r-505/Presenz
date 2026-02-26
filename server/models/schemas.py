# server/models/schemas.py

from pydantic import BaseModel, Field


class AttendanceRequest(BaseModel):
    """
    Request body for attendance submission.
    """

    name: str = Field(
        ...,
        min_length=2,
        max_length=100,
        description="Student full name",
    )

    roll: str = Field(
        ...,
        min_length=1,
        max_length=30,
        description="Student roll number",
    )

    session_code: str = Field(
        ...,
        min_length=1,
        max_length=32,
        description="Session verification code provided by teacher",
    )

    class Config:
        str_strip_whitespace = True