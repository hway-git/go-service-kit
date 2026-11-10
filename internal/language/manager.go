package language

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed lang/*.json
var fs embed.FS

type Manager struct {
	fs embed.FS
}

func NewManager() *Manager {
	return &Manager{fs: fs}
}

func (m *Manager) Load(lang string) (map[string]string, error) {
	data, err := m.fs.ReadFile(fmt.Sprintf("lang/%s.json", lang))
	if err != nil {
		return nil, err
	}
	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}
