package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ItemId struct {
	ID          string
	DisplayName string
}

const ItemIdFileName = "itemids.json"

type ItemIdDict struct {
	Map map[string]ItemId
}

func (i *ItemIdDict) Add(id, name string) *logError {
	if _, okay := i.Map[id]; okay {
		return &logError{
			errMessage: fmt.Sprintf("error. itemId %s %s already exists", id, i.Map[id]),
			logMessage: fmt.Sprintf("Error. ItemID %s %s already exists", id, i.Map[id]),
		}
	}

	i.Map[id] = ItemId{
		ID:          id,
		DisplayName: name,
	}

	return nil
}

func (i *ItemIdDict) Remove(id string) *logError {
	if _, okay := i.Map[id]; !okay {
		return &logError{
			errMessage: fmt.Sprintf("error. itemId %s not found", id),
			logMessage: fmt.Sprintf("Error. ItemID %s not found", id),
		}
	}

	delete(i.Map, id)
	return nil
}

func (i *ItemIdDict) GetAll() []ItemId {
	ids := []ItemId{}

	for _, val := range i.Map {
		ids = append(ids, val)
	}

	return ids
}

func (i *ItemIdDict) DisplayName(id string) (string, *logError) {
	itemID, okay := i.Map[id]
	if !okay {
		return "", &logError{
			errMessage: fmt.Sprintf("error. erx %s not found", id),
			logMessage: fmt.Sprintf("Error. ERx %s not found", id),
		}
	}

	return itemID.DisplayName, nil
}

func (i *ItemIdDict) Load(dataPath string) (error, *logResponder) {
	logger := logResponder{}
	logger.AddInfo("Loading ItemIDs")

	_, err := os.Stat(filepath.Join(dataPath, ItemIdFileName))
	if err != nil {
		if os.IsNotExist(err) {
			logger.AddInfo(fmt.Sprintf("%s not found. Attempting to create file", ItemIdFileName))
			file, err := os.OpenFile(filepath.Join(dataPath, ItemIdFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				logger.AddError(fmt.Sprintf("Error. Unable to create ItemID save file: %s", err.Error()))
				return fmt.Errorf("error. unable to create erx save file: %s", err.Error()), &logger
			}
			_, err = file.WriteString("{}")
			if err != nil {
				logger.AddError(fmt.Sprintf("Error. Unable to write ERx save file: %s", err.Error()))
				return fmt.Errorf("error. unable to write erx save file: %s", err.Error()), &logger
			}
			file.Close()
		}
	}

	//-- Load json data from save file
	data, err := os.ReadFile(filepath.Join(dataPath, ItemIdFileName))
	if err != nil {
		logger.AddError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err, &logger
	}

	//-- Unmarshal json into process memory
	err = json.Unmarshal(data, &i)
	if err != nil {
		logger.AddError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err, &logger

	}

	logger.AddInfo("ItemIDs loaded successfully")
	return nil, &logger
}

func (i *ItemIdDict) Save(dataPath string) (error, *logResponder) {
	logger := logResponder{}
	logger.AddInfo("Saving ItemIDs")

	file, err := os.OpenFile(filepath.Join(dataPath, ItemIdFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		logger.AddError(fmt.Sprintf("Error. Unable to open %s: %s", ERxFileName, err.Error()))
		return err, &logger
	}
	defer file.Close()

	data, err := json.Marshal(&i)
	if err != nil {
		logger.AddError(fmt.Sprintf("Error marshalling itemIDs: %s", err.Error()))
		return err, &logger
	}

	_, err = file.Write(data)
	if err != nil {
		logger.AddError(fmt.Sprintf("Error writting marshalled itemIDs to file %s: %s", ERxFileName, err.Error()))
		return err, &logger
	}

	return nil, &logger

}

func initItemIDs() *ItemIdDict {
	return &ItemIdDict{
		Map: map[string]ItemId{},
	}
}
