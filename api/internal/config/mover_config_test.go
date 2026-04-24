package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetMoverHooks(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		osGetwdFn = os.Getwd
		userHomeDirFn = os.UserHomeDir
		mkdirAllFn = os.MkdirAll
		osReadFileFn = os.ReadFile
		osWriteFileFn = os.WriteFile
		osRemoveFn = os.Remove
		osStatFn = os.Stat
		jsonMarshalIndentFn = json.MarshalIndent
		ServerURLProvider = nil
		ConfigFileNameProvider = nil
		AuthFileNameProvider = nil
		SyncStateFileNameProvider = nil
	})
}

func TestGetMoverConfigFileNames_withProviders(t *testing.T) {
	resetMoverHooks(t)
	ConfigFileNameProvider = func() string { return "custom.cfg" }
	AuthFileNameProvider = func() string { return "a.json" }
	SyncStateFileNameProvider = func() string { return "sync.json" }
	assert.Equal(t, "custom.cfg", getMoverConfigFileName())
	assert.Equal(t, "a.json", getMoverAuthFileName())
	assert.Equal(t, "sync.json", getMoverSyncStateFileName())
}

func TestGetServerURL(t *testing.T) {
	resetMoverHooks(t)
	assert.PanicsWithValue(t,
		"config: ServerURLProvider not set. CLI must call config.InitCLIServerURL at startup.",
		func() { GetServerURL() },
	)
	ServerURLProvider = func() string { return "https://api.example" }
	assert.Equal(t, "https://api.example", GetServerURL())
}

func TestGetConfigPath_getwdError(t *testing.T) {
	resetMoverHooks(t)
	osGetwdFn = func() (string, error) { return "", errors.New("no cwd") }
	_, err := getConfigPath()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get current directory")
}

func TestGetConfigPath_usesGitRoot_tChdir(t *testing.T) {
	resetMoverHooks(t)
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".git"), 0o700))
	sub := filepath.Join(root, "deep", "nest")
	require.NoError(t, os.MkdirAll(sub, 0o700))
	t.Chdir(sub)
	p, err := getConfigPath()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, ".nixopus"), p)
}

func TestGetConfigPath_noGitUsesCwd(t *testing.T) {
	resetMoverHooks(t)
	root := t.TempDir()
	t.Chdir(root)
	p, err := getConfigPath()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, ".nixopus"), p)
}

func TestLoad_getConfigPathError(t *testing.T) {
	resetMoverHooks(t)
	osGetwdFn = func() (string, error) { return "", errors.New("cwd") }
	ServerURLProvider = func() string { return "https://x" }
	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get current directory")
}

func TestLoad_configNotFound(t *testing.T) {
	resetMoverHooks(t)
	root := t.TempDir()
	t.Chdir(root)
	ServerURLProvider = func() string { return "https://x" }
	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "config not found")
}

func TestLoad_readFileError(t *testing.T) {
	resetMoverHooks(t)
	root := t.TempDir()
	t.Chdir(root)
	cfgPath := filepath.Join(root, ".nixopus")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`{"project_id":"p"}`), 0o600))
	ServerURLProvider = func() string { return "https://x" }
	osReadFileFn = func(name string) ([]byte, error) {
		if name == cfgPath {
			return nil, errors.New("read denied")
		}
		return os.ReadFile(name)
	}
	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestLoad_invalidJSON(t *testing.T) {
	resetMoverHooks(t)
	root := t.TempDir()
	t.Chdir(root)
	require.NoError(t, os.WriteFile(filepath.Join(root, ".nixopus"), []byte(`{`), 0o600))
	ServerURLProvider = func() string { return "https://x" }
	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse config file")
}

func TestLoad_noApplications(t *testing.T) {
	resetMoverHooks(t)
	root := t.TempDir()
	t.Chdir(root)
	require.NoError(t, os.WriteFile(filepath.Join(root, ".nixopus"), []byte(`{"sync":{"debounce_ms":1}}`), 0o600))
	ServerURLProvider = func() string { return "https://x" }
	_, err := Load()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no applications found")
}

func TestLoad_applicationsOnly(t *testing.T) {
	resetMoverHooks(t)
	root := t.TempDir()
	t.Chdir(root)
	raw := `{"applications":{"default":"app-id-1","web":"w2"}}`
	require.NoError(t, os.WriteFile(filepath.Join(root, ".nixopus"), []byte(raw), 0o600))
	ServerURLProvider = func() string { return "https://api" }
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "app-id-1", cfg.Applications["default"])
}

func TestLoad_statNonNotExistError(t *testing.T) {
	resetMoverHooks(t)
	root := t.TempDir()
	t.Chdir(root)
	cfgPath := filepath.Join(root, ".nixopus")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`{"project_id":"p"}`), 0o600))
	ServerURLProvider = func() string { return "https://x" }
	osStatFn = func(name string) (os.FileInfo, error) {
		if name == cfgPath {
			return nil, errors.New("stat err")
		}
		return os.Stat(name)
	}
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "p", cfg.ProjectID)
}

func TestLoad_successAndSyncDefaults(t *testing.T) {
	resetMoverHooks(t)
	root := t.TempDir()
	t.Chdir(root)
	raw := `{"project_id":"proj1","sync":{"debounce_ms":0,"exclude":[]}}`
	require.NoError(t, os.WriteFile(filepath.Join(root, ".nixopus"), []byte(raw), 0o600))
	ServerURLProvider = func() string { return "https://api.test" }
	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, "https://api.test", cfg.Server)
	assert.Equal(t, "proj1", cfg.ProjectID)
	assert.Equal(t, 300, cfg.Sync.DebounceMs)
	assert.Contains(t, cfg.Sync.Exclude, ".git")
}

func TestSave_errors(t *testing.T) {
	resetMoverHooks(t)
	root := t.TempDir()
	t.Chdir(root)
	ServerURLProvider = func() string { return "https://x" }
	c := &Config{ProjectID: "p1", Sync: SyncConfig{DebounceMs: 100}}

	jsonMarshalIndentFn = func(v any, prefix, indent string) ([]byte, error) {
		return nil, errors.New("marshal")
	}
	assert.ErrorContains(t, c.Save(), "failed to marshal config")

	jsonMarshalIndentFn = json.MarshalIndent
	osWriteFileFn = func(name string, data []byte, perm os.FileMode) error {
		return errors.New("write fail")
	}
	assert.ErrorContains(t, c.Save(), "failed to write config file")
}

func TestSave_getConfigPathError(t *testing.T) {
	resetMoverHooks(t)
	osGetwdFn = func() (string, error) { return "", errors.New("cwd") }
	c := &Config{ProjectID: "p"}
	err := c.Save()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get config path")
}

func TestSave_success(t *testing.T) {
	resetMoverHooks(t)
	root := t.TempDir()
	t.Chdir(root)
	ServerURLProvider = func() string { return "https://x" }
	c := &Config{
		ProjectID: "pid",
		Sync:      SyncConfig{DebounceMs: 50, Exclude: []string{"a"}},
		EnvPath:   "e",
	}
	require.NoError(t, c.Save())
	b, err := os.ReadFile(filepath.Join(root, ".nixopus"))
	require.NoError(t, err)
	var out map[string]any
	require.NoError(t, json.Unmarshal(b, &out))
	assert.Equal(t, "pid", out["project_id"])
}

func TestConfig_GetApplicationID(t *testing.T) {
	c := &Config{
		Applications: map[string]string{
			"default": "d1",
			"app":     "a1",
		},
	}
	id, err := c.GetApplicationID("")
	require.NoError(t, err)
	assert.Equal(t, "d1", id)
	id, err = c.GetApplicationID("app")
	require.NoError(t, err)
	assert.Equal(t, "a1", id)
	c2 := &Config{Applications: map[string]string{"app": "a1"}}
	_, err = c2.GetApplicationID("")
	require.Error(t, err)
	_, err = c.GetApplicationID("missing")
	require.Error(t, err)

	c3 := &Config{Applications: map[string]string{"default": ""}}
	_, err = c3.GetApplicationID("")
	require.Error(t, err)
}

func TestLoadAuth_getAuthPathError(t *testing.T) {
	resetMoverHooks(t)
	userHomeDirFn = func() (string, error) { return "", errors.New("no home") }
	_, err := LoadAuth()
	require.Error(t, err)
}

func TestClearAuth_getAuthPathError(t *testing.T) {
	resetMoverHooks(t)
	userHomeDirFn = func() (string, error) { return "", errors.New("no home") }
	assert.Error(t, ClearAuth())
}

func TestValidateEnvPath(t *testing.T) {
	resetMoverHooks(t)
	assert.NoError(t, ValidateEnvPath(""))
	assert.ErrorContains(t, ValidateEnvPath("/abs"), "absolute path")
	assert.ErrorContains(t, ValidateEnvPath("../x"), "path traversal")
	assert.ErrorContains(t, ValidateEnvPath("..\\x"), "path traversal")

	root := t.TempDir()
	t.Chdir(root)
	assert.ErrorContains(t, ValidateEnvPath("missing.env"), "env file not found")

	f := filepath.Join(root, "ok.env")
	require.NoError(t, os.WriteFile(f, []byte("x"), 0o600))
	assert.NoError(t, ValidateEnvPath("ok.env"))

	osGetwdFn = func() (string, error) { return "", errors.New("bad") }
	assert.ErrorContains(t, ValidateEnvPath("ok.env"), "failed to get current directory")
}

func TestGetAuthPath_userHomeError(t *testing.T) {
	resetMoverHooks(t)
	userHomeDirFn = func() (string, error) { return "", errors.New("no home") }
	_, err := getAuthPath()
	require.Error(t, err)
}

func TestEnsureAuthDir_mkdirError(t *testing.T) {
	resetMoverHooks(t)
	mkdirAllFn = func(path string, perm os.FileMode) error { return errors.New("mkdir") }
	assert.ErrorContains(t, ensureAuthDir(), "failed to create auth directory")
}

func TestGetSyncStatePath_errors(t *testing.T) {
	resetMoverHooks(t)
	mkdirAllFn = func(path string, perm os.FileMode) error { return errors.New("mkdir") }
	_, err := GetSyncStatePath()
	require.Error(t, err)

	home := t.TempDir()
	var homeCalls int
	mkdirAllFn = os.MkdirAll
	userHomeDirFn = func() (string, error) {
		homeCalls++
		if homeCalls == 1 {
			return home, nil
		}
		return "", errors.New("home")
	}
	_, err = GetSyncStatePath()
	require.Error(t, err)
}

func TestGetSyncStatePath_success(t *testing.T) {
	resetMoverHooks(t)
	home := t.TempDir()
	userHomeDirFn = func() (string, error) { return home, nil }
	p, err := GetSyncStatePath()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, ".config", "nixopus", "sync-state.json"), p)
}

func withFakeHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	userHomeDirFn = func() (string, error) { return home, nil }
	mkdirAllFn = os.MkdirAll
	return home
}

func TestLoadAuth_roundTrip(t *testing.T) {
	resetMoverHooks(t)
	_ = withFakeHome(t)
	a, err := LoadAuth()
	require.NoError(t, err)
	assert.Empty(t, a.AccessToken)

	require.NoError(t, SaveAuth("tok", "ref"))
	a2, err := LoadAuth()
	require.NoError(t, err)
	assert.Equal(t, "tok", a2.AccessToken)
	assert.Equal(t, "ref", a2.RefreshToken)

	require.NoError(t, SaveOrganizationID("org-1"))
	a3, err := LoadAuth()
	require.NoError(t, err)
	assert.Equal(t, "org-1", a3.OrganizationID)

	tok, err := GetAccessToken()
	require.NoError(t, err)
	assert.Equal(t, "tok", tok)

	oid, err := GetOrganizationID()
	require.NoError(t, err)
	assert.Equal(t, "org-1", oid)

	require.NoError(t, ClearAuth())
	a4, err := LoadAuth()
	require.NoError(t, err)
	assert.Empty(t, a4.AccessToken)
}

func TestLoadAuth_readError(t *testing.T) {
	resetMoverHooks(t)
	home := withFakeHome(t)
	dir := filepath.Join(home, ".config", "nixopus")
	require.NoError(t, os.MkdirAll(dir, 0o700))
	authPath := filepath.Join(dir, "auth.json")
	require.NoError(t, os.WriteFile(authPath, []byte(`{}`), 0o000))
	t.Cleanup(func() { _ = os.Chmod(authPath, 0o600) })
	_ = os.Chmod(authPath, 0o000)
	_, err := LoadAuth()
	if err == nil {
		// Some platforms allow root read; skip assert
		t.Skip("chmod did not make file unreadable")
	}
	assert.Contains(t, err.Error(), "failed to read auth file")
}

func TestLoadAuth_invalidJSON(t *testing.T) {
	resetMoverHooks(t)
	home := withFakeHome(t)
	dir := filepath.Join(home, ".config", "nixopus")
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "auth.json"), []byte(`{`), 0o600))
	_, err := LoadAuth()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse auth file")
}

func TestSaveAuth_errors(t *testing.T) {
	resetMoverHooks(t)
	mkdirAllFn = func(string, os.FileMode) error { return errors.New("mkdir") }
	assert.Error(t, SaveAuth("a", "b"))

	mkdirAllFn = os.MkdirAll
	userHomeDirFn = func() (string, error) { return "", errors.New("home") }
	assert.Error(t, SaveAuth("a", "b"))
}

func TestSaveAuth_getAuthPathError(t *testing.T) {
	resetMoverHooks(t)
	home := t.TempDir()
	var calls int
	userHomeDirFn = func() (string, error) {
		calls++
		if calls == 1 {
			return home, nil // ensureAuthDir
		}
		return "", errors.New("home later") // getAuthPath
	}
	assert.Error(t, SaveAuth("a", "b"))
}

func TestSaveOrganizationID_getAuthPathError(t *testing.T) {
	resetMoverHooks(t)
	home := t.TempDir()
	var calls int
	userHomeDirFn = func() (string, error) {
		calls++
		if calls <= 2 {
			return home, nil
		}
		return "", errors.New("home later")
	}
	assert.Error(t, SaveOrganizationID("org"))
}

func TestSaveAuth_preservesOrgWhenLoadAuthFails(t *testing.T) {
	resetMoverHooks(t)
	home := withFakeHome(t)
	dir := filepath.Join(home, ".config", "nixopus")
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "auth.json"), []byte(`{`), 0o600))
	require.NoError(t, SaveAuth("newtok", "newref"))
	a, err := LoadAuth()
	require.NoError(t, err)
	assert.Equal(t, "newtok", a.AccessToken)
	assert.Equal(t, "newref", a.RefreshToken)
}

func TestSaveAuth_marshalAndWriteErrors(t *testing.T) {
	resetMoverHooks(t)
	_ = withFakeHome(t)
	jsonMarshalIndentFn = func(any, string, string) ([]byte, error) { return nil, errors.New("m") }
	assert.ErrorContains(t, SaveAuth("x", "y"), "failed to marshal auth")

	jsonMarshalIndentFn = json.MarshalIndent
	osWriteFileFn = func(string, []byte, os.FileMode) error { return errors.New("w") }
	assert.ErrorContains(t, SaveAuth("x", "y"), "failed to write auth file")
}

func TestSaveOrganizationID_errors(t *testing.T) {
	resetMoverHooks(t)
	userHomeDirFn = func() (string, error) { return "", errors.New("h") }
	assert.Error(t, SaveOrganizationID("o"))

	home := withFakeHome(t)
	userHomeDirFn = func() (string, error) { return home, nil }
	jsonMarshalIndentFn = func(any, string, string) ([]byte, error) { return nil, errors.New("m") }
	assert.ErrorContains(t, SaveOrganizationID("o"), "failed to marshal auth")

	jsonMarshalIndentFn = json.MarshalIndent
	osWriteFileFn = func(string, []byte, os.FileMode) error { return errors.New("w") }
	assert.ErrorContains(t, SaveOrganizationID("o"), "failed to write auth file")
}

func TestGetAccessToken_and_GetOrganizationID_errors(t *testing.T) {
	resetMoverHooks(t)
	_ = withFakeHome(t)
	_, err := GetAccessToken()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not authenticated")

	_, err = GetOrganizationID()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no organization ID")

	require.NoError(t, SaveAuth("only-token", ""))
	_, err = GetOrganizationID()
	require.Error(t, err)
}

func TestGetAccessToken_and_GetOrganizationID_loadAuthFails(t *testing.T) {
	resetMoverHooks(t)
	home := withFakeHome(t)
	dir := filepath.Join(home, ".config", "nixopus")
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "auth.json"), []byte(`{`), 0o600))
	_, err := GetAccessToken()
	require.Error(t, err)
	_, err = GetOrganizationID()
	require.Error(t, err)
}

func TestSaveAuth_preservesOrganizationID(t *testing.T) {
	resetMoverHooks(t)
	_ = withFakeHome(t)
	require.NoError(t, SaveOrganizationID("org-keep"))
	require.NoError(t, SaveAuth("tok2", "ref2"))
	a, err := LoadAuth()
	require.NoError(t, err)
	assert.Equal(t, "org-keep", a.OrganizationID)
	assert.Equal(t, "tok2", a.AccessToken)
}

func TestSaveOrganizationID_whenLoadAuthFails(t *testing.T) {
	resetMoverHooks(t)
	home := withFakeHome(t)
	dir := filepath.Join(home, ".config", "nixopus")
	require.NoError(t, os.MkdirAll(dir, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "auth.json"), []byte(`{`), 0o600))
	require.NoError(t, SaveOrganizationID("org-new"))
	a, err := LoadAuth()
	require.NoError(t, err)
	assert.Equal(t, "org-new", a.OrganizationID)
}

func TestClearAuth_removeError(t *testing.T) {
	resetMoverHooks(t)
	home := withFakeHome(t)
	dir := filepath.Join(home, ".config", "nixopus")
	require.NoError(t, os.MkdirAll(dir, 0o700))
	p := filepath.Join(dir, "auth.json")
	require.NoError(t, os.WriteFile(p, []byte(`{}`), 0o600))
	osRemoveFn = func(string) error { return errors.New("rm") }
	assert.ErrorContains(t, ClearAuth(), "failed to remove auth file")
}

func TestClearAuth_noFile(t *testing.T) {
	resetMoverHooks(t)
	_ = withFakeHome(t)
	require.NoError(t, ClearAuth())
}
