package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const ERxFileName = "erxs.json"

type ERx struct {
	MedID       string
	DisplayName string
}

type ERxDict struct {
	Map map[string]ERx
}

func (e *ERxDict) Add(erx, name string) *logError {
	if _, okay := e.Map[erx]; okay {
		return &logError{
			errMessage: fmt.Sprintf("error. erx %s %s already exists.", erx, e.Map[erx]),
			logMessage: fmt.Sprintf("Error. ERx %s %s already exists.", erx, e.Map[erx]),
		}
	}

	e.Map[erx] = ERx{
		MedID:       erx,
		DisplayName: name,
	}

	return nil
}

func (e *ERxDict) Remove(erx string) *logError {
	if _, okay := e.Map[erx]; !okay {
		return &logError{
			errMessage: fmt.Sprintf("erx %s not found", erx),
			logMessage: fmt.Sprintf("ERx %s not found", erx),
		}
	}

	delete(e.Map, erx)
	return nil

}

func (e *ERxDict) GetAll() []ERx {
	erxs := []ERx{}
	for _, val := range e.Map {
		erxs = append(erxs, val)
	}

	return erxs
}

func (e *ERxDict) DisplayName(medID string) (string, *logError) {
	erx, okay := e.Map[medID]
	if !okay {
		return "", &logError{
			errMessage: fmt.Sprintf("error. erx %s not found", medID),
			logMessage: fmt.Sprintf("Error. ERx %s not found", medID),
		}
	}

	return erx.DisplayName, nil
}

func (e *ERxDict) Load(p *Process) error {
	p.logger.LogInfo("Loading ERXs")

	_, err := os.Stat(filepath.Join(p.pathToData, ERxFileName))
	if err != nil {
		if os.IsNotExist(err) {
			p.logger.LogInfo(fmt.Sprintf("%s not found. Attempting to create file", ERxFileName))
			file, err := os.OpenFile(filepath.Join(p.pathToData, ERxFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				p.logger.LogError(fmt.Sprintf("Error. Unable to create ERx save file: %s", err.Error()))
				return fmt.Errorf("error. unable to create erx save file: %s", err.Error())
			}
			_, err = file.WriteString("{}")
			if err != nil {
				p.logger.LogError(fmt.Sprintf("Error. Unable to write ERx save file: %s", err.Error()))
				return fmt.Errorf("error. unable to write erx save file: %s", err.Error())
			}
			file.Close()
		}
	}

	//-- Load json data from save file
	data, err := os.ReadFile(filepath.Join(p.pathToData, ERxFileName))
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err
	}

	//-- Unmarshal json into process memory
	err = json.Unmarshal(data, p.erxs)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err

	}

	p.logger.LogInfo("ERXs loaded successfully")
	p.state.ERXsLoadSuccessful()
	return nil

}

func (e *ERxDict) Save(p *Process) error {
	p.logger.LogInfo("Saving ERXs")

	file, err := os.OpenFile(filepath.Join(p.pathToData, ERxFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error. Unable to open %s: %s", ERxFileName, err.Error()))
		return err
	}
	defer file.Close()

	data, err := json.Marshal(p.erxs)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error marshalling ERXs: %s", err.Error()))
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error writting marshalled ERXs to file %s: %s", ERxFileName, err.Error()))
		return err
	}

	return nil
}

func initERXs() *ERxDict {
	return &ERxDict{
		Map: map[string]ERx{},
	}
}
