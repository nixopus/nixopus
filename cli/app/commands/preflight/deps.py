import shutil
import subprocess
from typing import Optional, Protocol

from pydantic import BaseModel, Field, field_validator

from app.utils.lib import ParallelProcessor, Supported
from app.utils.logger import Logger
from app.utils.output_formatter import OutputFormatter
from app.utils.protocols import LoggerProtocol

from .messages import (
    error_checking_dependency,
    invalid_os,
    invalid_package_manager,
    timeout_checking_dependency,
    debug_processing_deps,
    debug_dep_check_result,
    error_subprocess_execution_failed,
)


class DependencyCheckerProtocol(Protocol):
    def check_dependency(self, dep: str) -> bool: ...


class DependencyChecker:
    def __init__(self, logger: LoggerProtocol):
        self.logger = logger

    def check_dependency(self, dep: str) -> bool:
        try:
            is_available = shutil.which(dep) is not None
            self.logger.debug(debug_dep_check_result.format(dep=dep, status="available" if is_available else "not available"))
            return is_available

        except subprocess.TimeoutExpired:
            if self.logger.verbose:
                self.logger.error(timeout_checking_dependency.format(dep=dep))
            return False
        except Exception as e:
            if self.logger.verbose:
                self.logger.error(error_subprocess_execution_failed.format(dep=dep, error=e))
            return False


class DependencyValidator:
    def validate_os(self, os: str) -> str:
        if not Supported.os(os):
            raise ValueError(invalid_os.format(os=os))
        return os

    def validate_package_manager(self, package_manager: str) -> str:
        if not Supported.package_manager(package_manager):
            raise ValueError(invalid_package_manager.format(package_manager=package_manager))
        return package_manager


class DependencyFormatter:
    def __init__(self):
        self.output_formatter = OutputFormatter()

    def format_output(self, results: list["DepsCheckResult"], output: str) -> str:
        if not results:
            return self.output_formatter.format_output(
                self.output_formatter.create_success_message("No dependencies to check"), output
            )

        if len(results) == 1 and output == "text":
            messages = []
            result = results[0]
            message = f"{result.dependency} is {'available' if result.is_available else 'not available'}"
            if result.is_available:
                message = f"{result.dependency} is available"
                data = {"dependency": result.dependency, "is_available": result.is_available}
                messages.append(self.output_formatter.create_success_message(message, data))
            else:
                error = f"{result.dependency} is not available"
                data = {"dependency": result.dependency, "is_available": result.is_available, "error": result.error}
                messages.append(self.output_formatter.create_error_message(error, data))

        if output == "text":
            table_data = []
            for result in results:
                row = {
                    "Dependency": result.dependency,
                    "Status": "available" if result.is_available else "not available"
                }
                if result.error and not result.is_available:
                    row["Error"] = result.error
                table_data.append(row)
            
            return self.output_formatter.create_table(
                table_data,
                title="Dependency Check Results",
                show_header=True,
                show_lines=True
            )
        else:
            json_data = []
            for result in results:
                item = {
                    "dependency": result.dependency,
                    "is_available": result.is_available,
                    "status": "available" if result.is_available else "not available"
                }
                if result.error and not result.is_available:
                    item["error"] = result.error
                json_data.append(item)
            
            return self.output_formatter.format_json(json_data)


class DepsCheckResult(BaseModel):
    dependency: str
    verbose: bool
    output: str
    os: str
    package_manager: str
    is_available: bool = False
    error: Optional[str] = None


class DepsConfig(BaseModel):
    deps: list[str] = Field(..., min_length=1, description="The list of dependencies to check")
    verbose: bool = Field(False, description="Verbose output")
    output: str = Field("text", description="Output format, text, json")
    os: str = Field(..., description=f"The operating system to check, available: {Supported.get_os()}")
    package_manager: str = Field(..., description="The package manager to use")

    @field_validator("os")
    @classmethod
    def validate_os(cls, os: str) -> str:
        validator = DependencyValidator()
        return validator.validate_os(os)

    @field_validator("package_manager")
    @classmethod
    def validate_package_manager(cls, package_manager: str) -> str:
        validator = DependencyValidator()
        return validator.validate_package_manager(package_manager)


class DepsService:
    def __init__(self, config: DepsConfig, logger: LoggerProtocol = None, checker: DependencyCheckerProtocol = None):
        self.config = config
        self.logger = logger or Logger(verbose=config.verbose)
        self.checker = checker or DependencyChecker(self.logger)
        self.formatter = DependencyFormatter()

    def _create_result(self, dep: str, is_available: bool, error: str = None) -> DepsCheckResult:
        return DepsCheckResult(
            dependency=dep,
            verbose=self.config.verbose,
            output=self.config.output,
            os=self.config.os,
            package_manager=self.config.package_manager,
            is_available=is_available,
            error=error,
        )

    def _check_dependency(self, dep: str) -> DepsCheckResult:
        try:
            is_available = self.checker.check_dependency(dep)
            return self._create_result(dep, is_available)
        except Exception as e:
            return self._create_result(dep, False, str(e))

    def check_dependencies(self) -> list[DepsCheckResult]:
        self.logger.debug(debug_processing_deps.format(count=len(self.config.deps)))

        def process_dep(dep: str) -> DepsCheckResult:
            return self._check_dependency(dep)

        def error_handler(dep: str, error: Exception) -> DepsCheckResult:
            if self.logger.verbose:
                self.logger.error(error_checking_dependency.format(dep=dep, error=error))
            return self._create_result(dep, False, str(error))

        results = ParallelProcessor.process_items(
            items=self.config.deps,
            processor_func=process_dep,
            max_workers=min(len(self.config.deps), 50),
            error_handler=error_handler,
        )

        return results

    def check_and_format(self) -> str:
        results = self.check_dependencies()
        return self.formatter.format_output(results, self.config.output)


class Deps:
    def __init__(self, logger: LoggerProtocol = None):
        self.logger = logger
        self.validator = DependencyValidator()
        self.formatter = DependencyFormatter()

    def check(self, config: DepsConfig) -> list[DepsCheckResult]:
        service = DepsService(config, logger=self.logger)
        return service.check_dependencies()

    def format_output(self, results: list[DepsCheckResult], output: str) -> str:
        return self.formatter.format_output(results, output)
