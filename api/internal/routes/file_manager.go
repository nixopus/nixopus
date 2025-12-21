package routes

import (
	"github.com/go-fuego/fuego"
	file_manager "github.com/raghavyuva/nixopus-api/internal/features/file-manager/controller"
)

// RegisterFileManagerRoutes registers file manager routes
func (router *Router) RegisterFileManagerRoutes(fileManagerGroup *fuego.Server, fileManagerController *file_manager.FileManagerController) {
	fuego.Get(fileManagerGroup, "", fileManagerController.ListFiles)
	fuego.Post(fileManagerGroup, "/create-directory", fileManagerController.CreateDirectory)
	fuego.Post(fileManagerGroup, "/move-directory", fileManagerController.MoveDirectory)
	fuego.Post(fileManagerGroup, "/copy-directory", fileManagerController.CopyDirectory)
	fuego.Post(fileManagerGroup, "/upload", fileManagerController.UploadFile)
	fuego.Delete(fileManagerGroup, "/delete-directory", fileManagerController.DeleteDirectory)
}
