package web

import (
	"git.xenonstack.com/util/continuous-security-backend/config"
	"git.xenonstack.com/util/continuous-security-backend/src/database"
	"git.xenonstack.com/util/continuous-security-backend/src/method"
)

func WorkspaceNameUpdate(email string) error {
	db := config.DB
	return db.Model(database.RequestInfo{}).Where("email=?", email).Update("workspace", method.ProjectNamebyEmail(email)).Error
}
