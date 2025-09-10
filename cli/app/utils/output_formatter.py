import json
from typing import Any, Dict, List, Optional, Tuple, Union

from pydantic import BaseModel
from rich.console import Console
from rich.table import Table


class OutputMessage(BaseModel):
    success: bool
    message: str
    data: Optional[Dict[str, Any]] = None
    error: Optional[str] = None


class OutputFormatter:
    def __init__(self, invalid_output_format_msg: str = "Invalid output format"):
        self.invalid_output_format_msg = invalid_output_format_msg
        self.console = Console()

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
            return json.dumps([item.model_dump() if hasattr(item, "model_dump") else item for item in result], indent=2)
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

    def create_table(
        self,
        data: Union[Dict[str, Any], List[Dict[str, Any]]],
        title: Optional[str] = None,
        headers: Optional[Union[Tuple[str, str], List[str]]] = None,
        show_header: bool = True,
        show_lines: bool = False,
        column_styles: Optional[List[str]] = None,
    ) -> str:
        if not data:
            return "No data to display"

        table = Table(show_header=show_header, show_lines=show_lines)

        if title:
            table.title = title

        if isinstance(data, dict):
            if headers is None:
                headers = ("Key", "Value")

            if isinstance(headers, list):
                headers = tuple(headers[:2])

            if column_styles is None:
                column_styles = ["cyan", "magenta"]

            table.add_column(headers[0], style=column_styles[0], no_wrap=True)
            table.add_column(headers[1], style=column_styles[1])

            for key, value in sorted(data.items()):
                table.add_row(str(key), str(value))

        elif isinstance(data, list) and data:
            if headers is None:
                headers = list(data[0].keys())
            elif isinstance(headers, tuple):
                headers = list(headers)

            if column_styles is None:
                column_styles = ["cyan", "magenta", "green", "yellow", "blue", "red"] * (len(headers) // 6 + 1)

            for i, header in enumerate(headers):
                style = column_styles[i] if i < len(column_styles) else "white"
                table.add_column(str(header), style=style)

            for row in data:
                row_data = [str(row.get(header, "")) for header in headers]
                table.add_row(*row_data)

        with self.console.capture() as capture:
            self.console.print(table)

        return capture.get()

    def format_table_output(
        self,
        data: Union[Dict[str, str], List[Dict[str, Any]]],
        output_format: str,
        success_message: str,
        title: Optional[str] = None,
        headers: Optional[Union[Tuple[str, str], List[str]]] = None,
    ) -> str:
        if output_format == "json":
            return self.format_json({"success": True, "message": success_message, "data": data})
        else:
            if not data:
                return "No data to display"

            return self.create_table(data, title, headers).strip()
