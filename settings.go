package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const settingsFileName = "settings.json"

type Settings struct {
	PrintLogsToCliIO bool
}

func (p *Process) loadSettings() error {
	p.settings = &Settings{}

	//-- Check if settings file exists
	p.logger.LogInfo("Checking for custom settings")
	if _, err := os.Stat(filepath.Join(p.pathToSettings, settingsFileName)); err != nil {
		if os.IsNotExist(err) {
			p.logger.LogInfo("No custom settings found")
			return nil
		} else {
			p.logger.LogError(fmt.Sprintf("Error. Unable to access %s: %s", settingsFileName, err.Error()))
			return err
		}
	}

	//-- Load json data from save file
	p.logger.LogInfo("Custom settings found. Loading settings")
	data, err := os.ReadFile(filepath.Join(p.pathToSettings, settingsFileName))
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err
	}

	//-- Unmarshal json into process memory
	err = json.Unmarshal(data, p.settings)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err
	}

	return nil
}
