package main

type AppData struct {
	ServiceProviderProperties map[string]interface{}
}

// AppDataManager .
type AppDataManager interface {
	Load() (AppData, error)
	Save(AppData) error
}
