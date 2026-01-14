package initialize

import (
	"siro-backend/pkg/setting"
	"siro-backend/pkg/utils"
)

// Initializes all necessary components
func Initialize() {
	utils.InitJWT()
	setting.ConnectDB()
}
