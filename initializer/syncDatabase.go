package initializer

import "363project/model"

func SyncDatabase() {
	// bikin tabel 'users' otomatis di MySQL
	DB.AutoMigrate(&model.User{})
}
