"""
Data models and configuration for the conflict command.
"""

from typing import Optional
from pydantic import BaseModel, Field


class ConflictCheckResult(BaseModel):
    """Result of a conflict check for a tool."""
    tool: str
    expected: Optional[str] = None
    current: Optional[str] = None
    status: str
    conflict: bool
    error: Optional[str] = None


class ConflictConfig(BaseModel):
    """Configuration for conflict checking."""
    config_file: Optional[str] = Field(
        None, description="Path to configuration file (defaults to built-in config)"
    )
    verbose: bool = Field(False, description="Verbose output")
    output: str = Field("text", description="Output format (text/json)")
