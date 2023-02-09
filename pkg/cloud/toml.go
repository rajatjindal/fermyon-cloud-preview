package cloud

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type SpinToml struct {
	Name string
}

func GetAppNameFromSpinToml() (string, error) {
	raw, err := os.ReadFile("spin.toml")
	if err != nil {
		return "", err
	}

	var appConfig SpinToml
	err = toml.Unmarshal(raw, &appConfig)
	if err != nil {
		return "", err
	}

	return appConfig.Name, nil
}
