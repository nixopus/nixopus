import subprocess
from typing import Protocol, Optional
from pydantic import BaseModel, Field, field_validator
from app.utils.lib import Supported, ParallelProcessor
from app.utils.logger import Logger
from app.utils.protocols import LoggerProtocol
from app.utils.output_formatter import OutputFormatter
from .messages import invalid_os, invalid_package_manager, error_checking_dependency, timeout_checking_dependency

class DependencyCheckerProtocol(Protocol):
    def check_dependency(self, dep: str) -> bool:
        ...

class DependencyChecker:
    def __init__(self, timeout: int, logger: LoggerProtocol):
        self.timeout = timeout
        self.logger = logger
    
    def check_dependency(self, dep: str) -> bool:
        self.logger.debug(f"Checking dependency: {dep}")
        
        try:
            result = subprocess.run(
                ["command", "-v", dep],
                capture_output=True,
                text=True,
                timeout=self.timeout
            )
            return result.returncode == 0
            
        except subprocess.TimeoutExpired:
            self.logger.error(timeout_checking_dependency.format(dep=dep))
            return False
        except Exception as e:
            self.logger.error(error_checking_dependency.format(dep=dep, error=e))
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
                self.output_formatter.create_success_message("No dependencies to check"), 
                output
            )
        
        messages = []
        for result in results:
            if result.is_available:
                message = f"{result.dependency} is available"
                messages.append(self.output_formatter.create_success_message(message, result.model_dump()))
            else:
                error = f"{result.dependency} is not available"
                messages.append(self.output_formatter.create_error_message(error, result.model_dump()))
        
        return self.output_formatter.format_output(messages, output)

class DepsCheckResult(BaseModel):
    dependency:str
    timeout: int
    verbose: bool
    output: str
    os: str
    package_manager: str
    is_available: bool = False
    error: Optional[str] = None

class DepsConfig(BaseModel):
    deps: list[str] = Field(..., min_length=1, description="The list of dependencies to check")
    timeout: int = Field(1, gt=0, le=60, description="The timeout in seconds")
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
        self.checker = checker or DependencyChecker(config.timeout, self.logger)
        self.formatter = DependencyFormatter()
    
    def _create_result(self, dep: str, is_available: bool, error: str = None) -> DepsCheckResult:
        return DepsCheckResult(
            dependency=dep,
            timeout=self.config.timeout,
            verbose=self.config.verbose,
            output=self.config.output,
            os=self.config.os,
            package_manager=self.config.package_manager,
            is_available=is_available,
            error=error
        )
    
    def _check_dependency(self, dep: str) -> DepsCheckResult:
        try:
            is_available = self.checker.check_dependency(dep)
            return self._create_result(dep, is_available)
        except Exception as e:
            return self._create_result(dep, False, str(e))
    
    def check_dependencies(self) -> list[DepsCheckResult]:
        self.logger.debug(f"Checking dependencies: {self.config.deps}")
        
        def process_dep(dep: str) -> DepsCheckResult:
            return self._check_dependency(dep)
        
        def error_handler(dep: str, error: Exception) -> DepsCheckResult:
            self.logger.error(error_checking_dependency.format(dep=dep, error=error))
            return self._create_result(dep, False, str(error))
        
        results = ParallelProcessor.process_items(
            items=self.config.deps,
            processor_func=process_dep,
            max_workers=min(len(self.config.deps), 50),
            error_handler=error_handler
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
    