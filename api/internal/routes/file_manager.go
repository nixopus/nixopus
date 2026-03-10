package routes

import (
	"github.com/go-fuego/fuego"
	file_manager "github.com/raghavyuva/nixopus-api/internal/features/file-manager/controller"
)

// RegisterFileManagerRoutes registers file manager routes
func (router *Router) RegisterFileManagerRoutes(fileManagerGroup *fuego.Server, fileManagerController *file_manager.FileManagerController) {
	fuego.Get(
		fileManagerGroup,
		"",
		fileManagerController.ListFiles,
		fuego.OptionSummary("List files"),
		fuego.OptionQuery("path", "Directory path to list", fuego.ParamRequired()),
	)
	fuego.Post(
		fileManagerGroup,
		"/create-directory",
		fileManagerController.CreateDirectory,
		fuego.OptionSummary("Create directory"),
	)
	fuego.Post(
		fileManagerGroup,
		"/move-directory",
		fileManagerController.MoveDirectory,
		fuego.OptionSummary("Move directory"),
	)
	fuego.Post(
		fileManagerGroup,
		"/copy-directory",
		fileManagerController.CopyDirectory,
		fuego.OptionSummary("Copy directory"),
	)
	fuego.Post(
		fileManagerGroup,
		"/upload",
		fileManagerController.UploadFile,
		fuego.OptionSummary("Upload file"),
	)
	fuego.Delete(
		fileManagerGroup,
		"/delete-directory",
		fileManagerController.DeleteDirectory,
		fuego.OptionSummary("Delete directory"),
	)
}
