package extension

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	extcontroller "github.com/nixopus/nixopus/api/internal/features/extension/controller"
	extservice "github.com/nixopus/nixopus/api/internal/features/extension/service"
	exttypes "github.com/nixopus/nixopus/api/internal/features/extension/types"
	"github.com/nixopus/nixopus/api/internal/testutils"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func tplMetadataYAML(extensionID, version string) string {
	return `metadata:
  id: ` + extensionID + `
  name: Tpl Ext
  description: Desc
  author: a
  icon: i
  category: Utility
  type: install
  version: ` + version + `
  isVerified: false
  featured: false
`
}

func tplWriteTree(t *testing.T, root, folder, extensionID, version string) {
	t.Helper()
	sub := filepath.Join(root, folder)
	require.NoError(t, os.MkdirAll(sub, 0o700))
	path := filepath.Join(sub, "metadata.yaml")
	require.NoError(t, os.WriteFile(path, []byte(tplMetadataYAML(extensionID, version)), 0o600))
}

func sha256hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

func extensionFromMetadataFile(t *testing.T, filePath string) *types.Extension {
	t.Helper()
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)
	var y exttypes.ExtensionYAML
	require.NoError(t, yaml.Unmarshal(data, &y))
	parsed, err := json.Marshal(y)
	require.NoError(t, err)
	return &types.Extension{
		ExtensionID:      y.Metadata.ID,
		Name:             y.Metadata.Name,
		Description:      y.Metadata.Description,
		Author:           y.Metadata.Author,
		Icon:             y.Metadata.Icon,
		Category:         types.ExtensionCategory(y.Metadata.Category),
		ExtensionType:    types.ExtensionType(y.Metadata.Type),
		Version:          y.Metadata.Version,
		IsVerified:       y.Metadata.IsVerified,
		Featured:         y.Metadata.Featured,
		YAMLContent:      string(data),
		ParsedContent:    string(parsed),
		ContentHash:      sha256hex(string(data)),
		ValidationStatus: types.ValidationStatusValid,
	}
}

func newExtensionFuegoContext(t *testing.T, method, url string) fuego.ContextNoBody {
	t.Helper()
	req := httptest.NewRequest(method, url, nil)
	return fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
}

func seedTestExtension(t *testing.T, setup *testutils.TestSetup) (id uuid.UUID, extensionID string) {
	t.Helper()
	id = uuid.New()
	extensionID = "ctrl-int-" + strings.ReplaceAll(uuid.New().String(), "-", "")
	ext := &types.Extension{
		ID:               id,
		ExtensionID:      extensionID,
		Name:             "Integration Controller Ext",
		Description:      "d",
		Author:           "a",
		Icon:             "i",
		Category:         types.ExtensionCategoryUtility,
		ExtensionType:    types.ExtensionTypeInstall,
		Version:          "1.0.0",
		IsVerified:       false,
		Featured:         false,
		YAMLContent:      "y",
		ParsedContent:    `{}`,
		ContentHash:      "testhash-" + extensionID,
		ValidationStatus: types.ValidationStatusValid,
	}
	_, err := setup.DB.NewInsert().Model(ext).Exec(setup.Ctx)
	require.NoError(t, err)
	return id, extensionID
}

// --- TemplateLoader (real Postgres + filesystem)

func TestIntegration_NewTemplateLoader(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := extservice.NewTemplateLoader(setup.DB)
	require.NotNil(t, l)
}

func TestIntegration_TemplateLoader_LoadExtensionsFromDirectory_empty(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := extservice.NewTemplateLoader(setup.DB)
	require.NoError(t, l.LoadExtensionsFromDirectory(setup.Ctx, t.TempDir()))
}

func TestIntegration_TemplateLoader_LoadExtensionsFromDirectory_nonexistent(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := extservice.NewTemplateLoader(setup.DB)
	err := l.LoadExtensionsFromDirectory(setup.Ctx, "/nonexistent-template-dir-xyz")
	require.Error(t, err)
}

func TestIntegration_TemplateLoader_insert_skip_update_orphan_restore(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := extservice.NewTemplateLoader(setup.DB)

	extID := "svc-tpl-" + strings.ReplaceAll(uuid.New().String(), "-", "")
	orphanID := "svc-tpl-" + strings.ReplaceAll(uuid.New().String(), "-", "")

	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, "note.txt"), []byte("not a dir"), 0o600))
	tplWriteTree(t, root, "a", extID, "1.0.0")
	tplWriteTree(t, root, "keep", "svc-tpl-keep-"+strings.ReplaceAll(uuid.New().String(), "-", ""), "1.0.0")

	metaA := filepath.Join(root, "a", "metadata.yaml")
	extOut := extensionFromMetadataFile(t, metaA)
	orphan := *extOut
	orphan.ID = uuid.New()
	orphan.ExtensionID = orphanID
	orphan.YAMLContent = "x"
	orphan.ParsedContent = "{}"
	orphan.ContentHash = "deadbeeforphan"
	_, err := setup.DB.NewInsert().Model(&orphan).Exec(setup.Ctx)
	require.NoError(t, err)

	require.NoError(t, l.LoadExtensionsFromDirectory(setup.Ctx, root))
	require.NoError(t, l.LoadExtensionsFromDirectory(setup.Ctx, root))

	var got types.Extension
	err = setup.DB.NewSelect().Model(&got).Where("extension_id = ?", extID).Scan(setup.Ctx)
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", got.Version)

	tplWriteTree(t, root, "a", extID, "2.0.0")
	require.NoError(t, l.LoadExtensionsFromDirectory(setup.Ctx, root))
	err = setup.DB.NewSelect().Model(&got).Where("extension_id = ?", extID).Scan(setup.Ctx)
	require.NoError(t, err)
	assert.Equal(t, "2.0.0", got.Version)

	var orphanRow types.Extension
	require.NoError(t, setup.DB.NewSelect().Model(&orphanRow).Where("extension_id = ?", orphanID).Scan(setup.Ctx))
	require.NotNil(t, orphanRow.DeletedAt)

	extIDRestore := "svc-tpl-" + strings.ReplaceAll(uuid.New().String(), "-", "")
	tplWriteTree(t, root, "c", extIDRestore, "1.0.0")
	toRestore := extensionFromMetadataFile(t, filepath.Join(root, "c", "metadata.yaml"))
	toRestore.ID = uuid.New()
	del := time.Now().Add(-time.Hour)
	toRestore.DeletedAt = &del
	_, err = setup.DB.NewInsert().Model(toRestore).Exec(setup.Ctx)
	require.NoError(t, err)

	require.NoError(t, l.LoadExtensionsFromDirectory(setup.Ctx, root))
	err = setup.DB.NewSelect().Model(&got).Where("extension_id = ?", extIDRestore).Scan(setup.Ctx)
	require.NoError(t, err)
	assert.Nil(t, got.DeletedAt)
}

func TestIntegration_TemplateLoader_insert_noVariables(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := extservice.NewTemplateLoader(setup.DB)
	extID := "svc-tpl-novar-" + strings.ReplaceAll(uuid.New().String(), "-", "")[:20]
	root := t.TempDir()
	sub := filepath.Join(root, "nv")
	require.NoError(t, os.MkdirAll(sub, 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(sub, "metadata.yaml"), []byte(tplMetadataYAML(extID, "1.0.0")), 0o600))

	require.NoError(t, l.LoadExtensionsFromDirectory(setup.Ctx, root))
	got, err := l.GetExtensionByID(setup.Ctx, extID)
	require.NoError(t, err)
	assert.Empty(t, got.Variables)
}

func TestIntegration_TemplateLoader_GetExtensionByID_notFound(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := extservice.NewTemplateLoader(setup.DB)
	_, err := l.GetExtensionByID(setup.Ctx, "no-such-extension-id-zzzzz")
	require.Error(t, err)
}

func TestIntegration_TemplateLoader_LoadExtensionsFromTemplates(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := extservice.NewTemplateLoader(setup.DB)
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(wd) })

	root := t.TempDir()
	require.NoError(t, os.Chdir(root))
	require.NoError(t, os.MkdirAll("templates/sub", 0o700))
	extID := "svc-tpl-chdir-" + strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
	require.NoError(t, os.WriteFile("templates/sub/metadata.yaml", []byte(tplMetadataYAML(extID, "1.0.0")), 0o600))

	require.NoError(t, l.LoadExtensionsFromTemplates(setup.Ctx))
	got, err := l.GetExtensionByID(setup.Ctx, extID)
	require.NoError(t, err)
	assert.Equal(t, extID, got.ExtensionID)
}

func TestIntegration_TemplateLoader_secondLoad(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := extservice.NewTemplateLoader(setup.DB)
	root := t.TempDir()
	extID := "svc-tpl-rmnoop-" + strings.ReplaceAll(uuid.New().String(), "-", "")
	tplWriteTree(t, root, "one", extID, "1.0.0")
	require.NoError(t, l.LoadExtensionsFromDirectory(setup.Ctx, root))
	require.NoError(t, l.LoadExtensionsFromDirectory(setup.Ctx, root))
}

func TestIntegration_TemplateLoader_LoadExtensions_cancelledContext(t *testing.T) {
	setup := testutils.NewTestSetup()
	l := extservice.NewTemplateLoader(setup.DB)
	extID := "svc-tpl-cancel-" + strings.ReplaceAll(uuid.New().String(), "-", "")[:16]
	root := t.TempDir()
	tplWriteTree(t, root, "x", extID, "1.0.0")

	ctx, cancel := context.WithCancel(setup.Ctx)
	cancel()
	err := l.LoadExtensionsFromDirectory(ctx, root)
	require.Error(t, err)
}

// --- Extensions HTTP controller (handlers + DB)

func TestIntegration_ExtensionsController_GetExtensions(t *testing.T) {
	setup := testutils.NewTestSetup()
	_, extensionID := seedTestExtension(t, setup)

	ctrl := extcontroller.NewExtensionsController(setup.Store, setup.Ctx, setup.Logger, nil)

	t.Run("list_default", func(t *testing.T) {
		ctx := newExtensionFuegoContext(t, http.MethodGet, "/v1/extensions")
		resp, err := ctrl.GetExtensions(ctx)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "success", resp.Status)
		found := false
		for _, e := range resp.Data.Extensions {
			if e.ExtensionID == extensionID {
				found = true
				break
			}
		}
		assert.True(t, found, "seeded extension should appear in list")
	})

	t.Run("list_filter_category", func(t *testing.T) {
		ctx := newExtensionFuegoContext(t, http.MethodGet, "/v1/extensions?category=Utility")
		resp, err := ctrl.GetExtensions(ctx)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.GreaterOrEqual(t, resp.Data.Total, 1)
	})
}

func TestIntegration_ExtensionsController_GetCategories(t *testing.T) {
	setup := testutils.NewTestSetup()
	seedTestExtension(t, setup)

	ctrl := extcontroller.NewExtensionsController(setup.Store, setup.Ctx, setup.Logger, nil)
	ctx := newExtensionFuegoContext(t, http.MethodGet, "/v1/extensions/categories")
	resp, err := ctrl.GetCategories(ctx)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "success", resp.Status)
	assert.GreaterOrEqual(t, len(resp.Data), 1)
}

func TestIntegration_ExtensionsController_GetExtension(t *testing.T) {
	setup := testutils.NewTestSetup()
	id, _ := seedTestExtension(t, setup)
	ctrl := extcontroller.NewExtensionsController(setup.Store, setup.Ctx, setup.Logger, nil)

	t.Run("ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/extensions/"+id.String(), nil)
		req.SetPathValue("id", id.String())
		ctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
		resp, err := ctrl.GetExtension(ctx)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, id, resp.Data.ID)
	})

	t.Run("missing_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/extensions/", nil)
		ctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
		_, err := ctrl.GetExtension(ctx)
		var b fuego.BadRequestError
		require.ErrorAs(t, err, &b)
	})

	t.Run("not_found", func(t *testing.T) {
		missing := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/v1/extensions/"+missing.String(), nil)
		req.SetPathValue("id", missing.String())
		ctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
		_, err := ctrl.GetExtension(ctx)
		var n fuego.NotFoundError
		require.ErrorAs(t, err, &n)
	})
}

func TestIntegration_ExtensionsController_GetExtensionByExtensionID(t *testing.T) {
	setup := testutils.NewTestSetup()
	_, extensionID := seedTestExtension(t, setup)
	ctrl := extcontroller.NewExtensionsController(setup.Store, setup.Ctx, setup.Logger, nil)

	t.Run("ok", func(t *testing.T) {
		url := "/v1/extensions/by-extension-id/" + extensionID
		req := httptest.NewRequest(http.MethodGet, url, nil)
		req.SetPathValue("extension_id", extensionID)
		ctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
		resp, err := ctrl.GetExtensionByExtensionID(ctx)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "success", resp.Status)
		assert.Equal(t, extensionID, resp.Data.ExtensionID)
	})

	t.Run("missing_extension_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/extensions/by-extension-id/", nil)
		ctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
		_, err := ctrl.GetExtensionByExtensionID(ctx)
		var b fuego.BadRequestError
		require.ErrorAs(t, err, &b)
	})

	t.Run("not_found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/extensions/by-extension-id/none-such-id", nil)
		req.SetPathValue("extension_id", "none-such-id")
		ctx := fuego.NewNetHTTPContext[any, any](fuego.BaseRoute{}, httptest.NewRecorder(), req, fuego.ReadOptions)
		_, err := ctrl.GetExtensionByExtensionID(ctx)
		var n fuego.NotFoundError
		require.ErrorAs(t, err, &n)
	})
}

func TestIntegration_NewExtensionsController(t *testing.T) {
	setup := testutils.NewTestSetup()
	ctrl := extcontroller.NewExtensionsController(setup.Store, setup.Ctx, setup.Logger, nil)
	require.NotNil(t, ctrl)
}
