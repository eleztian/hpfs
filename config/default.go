package config

import (
	"os"
)

// defaultConfigFileContent is default config json
const defaultConfigFileContent = `
{
    "git_path": ".git",
    "host_url": "http://127.0.0.1:8000",
    "share_paths": [
    ],
    "user": {
        "username": "tab.zhang",
        "password": "tab"
    },
    "rpc_port": "1234"
}
`

func createDefaultFile(filePath string) error {
	f, err := os.Create(filePath)
	// f, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(defaultConfigFileContent)
	if err != nil {
		return err
	}
	return nil
}
