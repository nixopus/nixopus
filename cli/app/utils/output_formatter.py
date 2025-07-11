import json
from typing import Any, Dict, Optional
from pydantic import BaseModel

class OutputMessage(BaseModel):
    success: bool
    message: str
    data: Optional[Dict[str, Any]] = None
    error: Optional[str] = None

class OutputFormatter:
    def __init__(self, invalid_output_format_msg: str = "Invalid output format"):
        self.invalid_output_format_msg = invalid_output_format_msg
    
    def format_text(self, result: Any) -> str:
        if isinstance(result, OutputMessage):
            if result.success:
                return result.message
            else:
                return f"Error: {result.error or 'Unknown error'}"
        elif isinstance(result, list):
            return "\n".join([self.format_text(item) for item in result])
        else:
            return str(result)
    
    def format_json(self, result: Any) -> str:
        if isinstance(result, OutputMessage):
            return json.dumps(result.model_dump(), indent=2)
        elif isinstance(result, list):
            return json.dumps([item.model_dump() if hasattr(item, 'model_dump') else item for item in result], indent=2)
        elif isinstance(result, BaseModel):
            return json.dumps(result.model_dump(), indent=2)
        else:
            return json.dumps(result, indent=2)
    
    def format_output(self, result: Any, output: str) -> str:
        if output == "text":
            return self.format_text(result)
        elif output == "json":
            return self.format_json(result)
        else:
            raise ValueError(self.invalid_output_format_msg)
    
    def create_success_message(self, message: str, data: Optional[Dict[str, Any]] = None) -> OutputMessage:
        return OutputMessage(success=True, message=message, data=data)
    
    def create_error_message(self, error: str, data: Optional[Dict[str, Any]] = None) -> OutputMessage:
        return OutputMessage(success=False, message="", error=error, data=data) 
