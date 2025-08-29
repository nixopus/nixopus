import subprocess
from unittest.mock import Mock, patch

import pytest
from pydantic import ValidationError

from app.commands.clone.clone import (
    Clone,
    CloneConfig,
    CloneFormatter,
    CloneResult,
    CloneService,
    GitClone,
    GitCommandBuilder,
)
from app.commands.clone.messages import (
    successfully_cloned,
    dry_run_mode,
    dry_run_command,
    dry_run_force_mode,
    path_exists_will_overwrite,
    path_exists_would_fail,
)
from app.utils.lib import DirectoryManager
from app.utils.logger import Logger


class TestGitCommandBuilder:
    def test_build_clone_command_without_branch(self):
        cmd = GitCommandBuilder.build_clone_command("https://github.com/user/repo", "/path/to/clone")
        assert cmd == ["git", "clone", "--depth=1", "https://github.com/user/repo", "/path/to/clone"]

    def test_build_clone_command_with_branch(self):
        cmd = GitCommandBuilder.build_clone_command("https://github.com/user/repo", "/path/to/clone", "main")
        assert cmd == ["git", "clone", "--depth=1", "-b", "main", "https://github.com/user/repo", "/path/to/clone"]

    def test_build_clone_command_with_empty_branch(self):
        cmd = GitCommandBuilder.build_clone_command("https://github.com/user/repo", "/path/to/clone", "")
        assert cmd == ["git", "clone", "--depth=1", "https://github.com/user/repo", "/path/to/clone"]


class TestCloneFormatter:
    def setup_method(self):
        self.formatter = CloneFormatter()

    def test_format_output_success(self):
        result = CloneResult(
            repo="https://github.com/user/repo",
            path="/path/to/clone",
            branch="main",
            force=False,
            verbose=False,
            output="text",
            success=True,
        )
        formatted = self.formatter.format_output(result, "text")
        assert successfully_cloned.format(repo="https://github.com/user/repo", path="/path/to/clone") in formatted

    def test_format_output_failure(self):
        result = CloneResult(
            repo="https://github.com/user/repo",
            path="/path/to/clone",
            branch="main",
            force=False,
            verbose=False,
            output="text",
            success=False,
            error="Repository not found",
        )
        formatted = self.formatter.format_output(result, "text")
        assert "Error: Repository not found" in formatted

    def test_format_output_json(self):
        result = CloneResult(
            repo="https://github.com/user/repo",
            path="/path/to/clone",
            branch="main",
            force=False,
            verbose=False,
            output="json",
            success=True,
        )
        formatted = self.formatter.format_output(result, "json")
        import json

        data = json.loads(formatted)
        assert data["success"] is True
        assert data["message"] == successfully_cloned.format(repo="https://github.com/user/repo", path="/path/to/clone")

    def test_format_output_invalid(self):
        result = CloneResult(
            repo="https://github.com/user/repo",
            path="/path/to/clone",
            branch="main",
            force=False,
            verbose=False,
            output="invalid",
            success=True,
        )
        with pytest.raises(ValueError):
            self.formatter.format_output(result, "invalid")

    @patch("os.path.exists")
    def test_format_dry_run(self, mock_exists):
        mock_exists.return_value = False
        config = CloneConfig(
            repo="https://github.com/user/repo", path="/path/to/clone", branch="main", force=True, dry_run=True
        )
        formatted = self.formatter.format_dry_run(config)
        assert dry_run_mode in formatted
        assert (
            dry_run_command.format(command="git clone --depth=1 -b main https://github.com/user/repo /path/to/clone")
            in formatted
        )
        assert dry_run_force_mode.format(force=True) in formatted

    @patch("os.path.exists")
    def test_format_dry_run_path_exists_force(self, mock_exists):
        mock_exists.return_value = True
        config = CloneConfig(
            repo="https://github.com/user/repo", path="/path/to/clone", branch="main", force=True, dry_run=True
        )
        formatted = self.formatter.format_dry_run(config)
        assert path_exists_will_overwrite.format(path="/path/to/clone") in formatted

    @patch("os.path.exists")
    def test_format_dry_run_path_exists_no_force(self, mock_exists):
        mock_exists.return_value = True
        config = CloneConfig(
            repo="https://github.com/user/repo", path="/path/to/clone", branch="main", force=False, dry_run=True
        )
        formatted = self.formatter.format_dry_run(config)
        assert path_exists_would_fail.format(path="/path/to/clone") in formatted


class TestGitClone:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.git_clone = GitClone(self.logger)

    @patch("subprocess.run")
    def test_clone_repository_success(self, mock_run):
        mock_run.return_value = Mock(returncode=0)

        success, error = self.git_clone.clone_repository("https://github.com/user/repo", "/path/to/clone", "main")

        assert success is True
        assert error is None
        self.logger.debug.assert_called()

    @patch("subprocess.run")
    def test_clone_repository_without_branch(self, mock_run):
        mock_run.return_value = Mock(returncode=0)

        success, error = self.git_clone.clone_repository("https://github.com/user/repo", "/path/to/clone")

        assert success is True
        assert error is None
        mock_run.assert_called_once()
        cmd = mock_run.call_args[0][0]
        assert cmd == ["git", "clone", "--depth=1", "https://github.com/user/repo", "/path/to/clone"]

    @patch("subprocess.run")
    def test_clone_repository_failure(self, mock_run):
        mock_run.side_effect = subprocess.CalledProcessError(1, "git clone", stderr="Repository not found")

        success, error = self.git_clone.clone_repository("https://github.com/user/repo", "/path/to/clone")

        assert success is False
        assert error == "Repository not found"

    @patch("subprocess.run")
    def test_clone_repository_unexpected_error(self, mock_run):
        mock_run.side_effect = Exception("Unexpected error")

        success, error = self.git_clone.clone_repository("https://github.com/user/repo", "/path/to/clone")

        assert success is False
        assert error == "Unexpected error"


class TestCloneConfig:
    def test_valid_config(self):
        config = CloneConfig(repo="https://github.com/user/repo", path="/path/to/clone", branch="main")
        assert config.repo == "https://github.com/user/repo"
        assert config.path == "/path/to/clone"
        assert config.branch == "main"
        assert config.force is False
        assert config.verbose is False
        assert config.output == "text"
        assert config.dry_run is False

    def test_valid_repo_formats(self):
        valid_repos = [
            "https://github.com/user/repo",
            "http://github.com/user/repo",
            "git://github.com/user/repo",
            "ssh://github.com/user/repo",
            "git@github.com:user/repo.git",
            "https://github.com/user/repo.git",
        ]

        for repo in valid_repos:
            config = CloneConfig(repo=repo, path="/path/to/clone")
            assert config.repo == repo

    def test_invalid_repo_formats(self):
        invalid_repos = ["", "   ", "github.com:user/repo", "invalid://github.com/user/repo"]

        for repo in invalid_repos:
            with pytest.raises(ValidationError):
                CloneConfig(repo=repo, path="/path/to/clone")

    def test_empty_repo(self):
        with pytest.raises(ValidationError):
            CloneConfig(repo="", path="/path/to/clone")

    def test_empty_path(self):
        with pytest.raises(ValidationError):
            CloneConfig(repo="https://github.com/user/repo", path="")

    def test_branch_validation(self):
        config = CloneConfig(repo="https://github.com/user/repo", path="/path/to/clone", branch="   ")
        assert config.branch is None

    def test_is_valid_repo_format(self):
        valid_repos = [
            "https://github.com/user/repo",
            "http://github.com/user/repo",
            "git@github.com:user/repo.git",
            "https://github.com/user/repo.git",
        ]

        for repo in valid_repos:
            assert CloneConfig._is_valid_repo_format(repo) is True

        invalid_repos = ["github.com:user/repo", "invalid://github.com/user/repo"]

        for repo in invalid_repos:
            assert CloneConfig._is_valid_repo_format(repo) is False


class TestDirectoryManager:
    def setup_method(self):
        self.logger = Mock(spec=Logger)

    @patch("shutil.rmtree")
    def test_remove_directory_success(self, mock_rmtree):
        success = DirectoryManager.remove_directory("/path/to/remove", self.logger)

        assert success is True
        mock_rmtree.assert_called_once_with("/path/to/remove")
        self.logger.debug.assert_called()

    @patch("shutil.rmtree")
    def test_remove_directory_failure(self, mock_rmtree):
        mock_rmtree.side_effect = Exception("Permission denied")

        success = DirectoryManager.remove_directory("/path/to/remove", self.logger)

        assert success is False
        self.logger.error.assert_called_once()

    @patch("os.path.exists")
    def test_path_exists_and_not_force_true(self, mock_exists):
        mock_exists.return_value = True

        result = DirectoryManager.path_exists_and_not_force("/path/to/check", False)

        assert result is True

    @patch("os.path.exists")
    def test_path_exists_and_not_force_false_when_force(self, mock_exists):
        mock_exists.return_value = True

        result = DirectoryManager.path_exists_and_not_force("/path/to/check", True)

        assert result is False

    @patch("os.path.exists")
    def test_path_exists_and_not_force_false_when_not_exists(self, mock_exists):
        mock_exists.return_value = False

        result = DirectoryManager.path_exists_and_not_force("/path/to/check", False)

        assert result is False


class TestCloneService:
    def setup_method(self):
        self.config = CloneConfig(repo="https://github.com/user/repo", path="/path/to/clone", branch="main")
        self.logger = Mock(spec=Logger)
        self.cloner = Mock(spec=GitClone)
        self.service = CloneService(self.config, self.logger, self.cloner)

    def test_create_result_success(self):
        result = self.service._create_result(True)

        assert result.repo == self.config.repo
        assert result.path == self.config.path
        assert result.branch == self.config.branch
        assert result.success is True
        assert result.error is None

    def test_create_result_failure(self):
        result = self.service._create_result(False, "Test error")

        assert result.success is False
        assert result.error == "Test error"

    @patch("os.path.exists")
    def test_validate_prerequisites_success(self, mock_exists):
        mock_exists.return_value = False

        result = self.service._validate_prerequisites()

        assert result is True

    @patch("os.path.exists")
    def test_validate_prerequisites_path_exists_no_force(self, mock_exists):
        mock_exists.return_value = True

        result = self.service._validate_prerequisites()

        assert result is False
        self.logger.error.assert_called_once()

    @patch("os.path.exists")
    def test_prepare_target_directory_force_success(self, mock_exists):
        self.service.config.force = True
        mock_exists.return_value = True
        self.service.dir_manager.remove_directory = Mock(return_value=True)

        result = self.service._prepare_target_directory()

        assert result is True
        self.service.dir_manager.remove_directory.assert_called_once_with(self.config.path, self.logger)

    @patch("os.path.exists")
    def test_prepare_target_directory_force_failure(self, mock_exists):
        self.service.config.force = True
        mock_exists.return_value = True
        self.service.dir_manager.remove_directory = Mock(return_value=False)

        result = self.service._prepare_target_directory()

        assert result is False
        self.service.dir_manager.remove_directory.assert_called_once_with(self.config.path, self.logger)

    def test_clone_success(self):
        self.cloner.clone_repository.return_value = (True, None)

        result = self.service.clone()

        assert result.success is True
        self.cloner.clone_repository.assert_called_once_with(self.config.repo, self.config.path, self.config.branch)

    def test_clone_failure(self):
        self.cloner.clone_repository.return_value = (False, "Test error")

        result = self.service.clone()

        assert result.success is False
        assert result.error == "Test error"

    def test_clone_and_format_dry_run(self):
        self.config.dry_run = True

        result = self.service.clone_and_format()

        assert dry_run_mode in result

    def test_clone_and_format_success(self):
        self.cloner.clone_repository.return_value = (True, None)

        result = self.service.clone_and_format()

        assert successfully_cloned.format(repo=self.config.repo, path=self.config.path) in result


class TestClone:
    def setup_method(self):
        self.logger = Mock(spec=Logger)
        self.clone = Clone(self.logger)

    def test_clone_success(self):
        config = CloneConfig(repo="https://github.com/user/repo", path="/path/to/clone", branch="main")

        with patch.object(CloneService, "clone") as mock_clone:
            mock_result = CloneResult(
                repo=config.repo,
                path=config.path,
                branch=config.branch,
                force=config.force,
                verbose=config.verbose,
                output=config.output,
                success=True,
            )
            mock_clone.return_value = mock_result

            result = self.clone.clone(config)

            assert result.success is True

    def test_format_output(self):
        result = CloneResult(
            repo="https://github.com/user/repo",
            path="/path/to/clone",
            branch="main",
            force=False,
            verbose=False,
            output="text",
            success=True,
        )

        formatted = self.clone.format_output(result, "text")

        assert successfully_cloned.format(repo="https://github.com/user/repo", path="/path/to/clone") in formatted

    def test_clone_and_format(self):
        config = CloneConfig(
            repo="https://github.com/user/repo",
            path="/path/to/clone",
            branch="main",
            force=False,
            verbose=False,
            output="text",
            dry_run=True,
        )

        with patch.object(CloneService, "clone_and_format") as mock_clone_and_format:
            mock_clone_and_format.return_value = dry_run_mode

            formatted = self.clone.clone_and_format(config)

            assert dry_run_mode in formatted

    def test_debug_logging_enabled(self):
        """Test that debug logging is properly enabled when verbose=True"""
        config = CloneConfig(
            repo="https://github.com/user/repo",
            path="/path/to/clone",
            branch="main",
            force=False,
            verbose=True,
            output="text",
            dry_run=False,
        )

        logger = Mock(spec=Logger)
        clone_operation = Clone(logger=logger)

        # Patch only GitClone.clone_repository to simulate a successful clone
        with patch("app.commands.clone.clone.GitClone.clone_repository", return_value=(True, None)):
            result = clone_operation.clone(config)

            # Verify that debug logging was called
            assert logger.debug.called
