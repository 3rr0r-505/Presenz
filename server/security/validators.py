# server/security/validators.py

import re


class ValidationError(Exception):
    """
    Raised when user input validation fails.
    """
    pass


# -------------------------------------------------
# BASIC SANITIZATION
# -------------------------------------------------
def sanitize_text(value: str) -> str:
    """
    Trim whitespace and remove control characters.
    """
    if not isinstance(value, str):
        raise ValidationError("Invalid input type")

    value = value.strip()

    # Remove ASCII control characters
    value = re.sub(r"[\x00-\x1f\x7f]", "", value)

    return value


# -------------------------------------------------
# NAME VALIDATION
# -------------------------------------------------
def validate_name(name: str) -> str:
    """
    Validate student name.
    Allowed:
        - Letters (A–Z, a–z)
        - Spaces
        - Dot (.)
    """
    name = sanitize_text(name)

    if not (2 <= len(name) <= 100):
        raise ValidationError("Name length must be between 2 and 100 characters")

    if not re.fullmatch(r"[A-Za-z.\s]+", name):
        raise ValidationError("Name contains invalid characters")

    return name


# -------------------------------------------------
# ROLL NUMBER VALIDATION
# -------------------------------------------------
def validate_roll(roll: str) -> str:
    """
    Validate roll number.
    Allowed:
        - Alphanumeric only
        - No spaces
        - No special characters
    """
    roll = sanitize_text(roll)

    if not (1 <= len(roll) <= 30):
        raise ValidationError("Roll number length must be between 1 and 30 characters")

    if not re.fullmatch(r"[A-Za-z0-9\-]+", roll):
        raise ValidationError("Roll number must be alphanumeric only")

    return roll.upper()


# -------------------------------------------------
# SESSION CODE VALIDATION
# -------------------------------------------------
def validate_session_code(code: str, expected_code: str) -> None:
    """
    Validate session code entered by student.
    The code is NOT stored in DB.
    """
    code = sanitize_text(code)

    if code != expected_code:
        raise ValidationError("Invalid session code")