package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const ERxItemIdLinksFileName = "ERxItemIdLinks.json"

type ERxItemIdLink struct {
	ERx    string
	ItemID string
}

type ERxItemIdLinks struct {
	Map map[ERxItemIdLink]struct{}
}

type ERxItemIdLinkSaveFormat struct {
	Map map[string]ERxItemIdLink
}

func (e *ERxItemIdLinks) Add(erx, itemId string) *logError {
	newLink := ERxItemIdLink{
		ERx:    erx,
		ItemID: itemId,
	}

	if _, okay := e.Map[newLink]; okay {
		return &logError{
			errMessage: fmt.Sprintf("error. erx %s and itemID %s are already linked", erx, itemId),
			logMessage: fmt.Sprintf("Error. ERx %s and ItemID %s are already linked", erx, itemId),
		}
	}

	e.Map[newLink] = struct{}{}

	return nil
}

func (e *ERxItemIdLinks) Remove(erx, itemId string) *logError {
	link := ERxItemIdLink{
		ERx:    erx,
		ItemID: itemId,
	}

	if _, okay := e.Map[link]; !okay {
		return &logError{
			errMessage: fmt.Sprintf("error. erx %s and itemId %s are not currently linked", erx, itemId),
			logMessage: fmt.Sprintf("Error. ERx %s and ItemId %s are not currently linked", erx, itemId),
		}
	}

	delete(e.Map, link)
	return nil
}

func (e *ERxItemIdLinks) GetItemIds(erx string) ([]string, *logError) {
	itemIDs := []string{}
	for link := range e.Map {
		if link.ERx == erx {
			itemIDs = append(itemIDs, link.ItemID)
		}
	}
	if len(itemIDs) == 0 {
		return nil, &logError{
			errMessage: fmt.Sprintf("error. no itemIDs linked to erx %s", erx),
			logMessage: fmt.Sprintf("Error. No itemIDs linked to ERx %s", erx),
		}
	}

	return itemIDs, nil
}

func (e *ERxItemIdLinks) GetMedIds(itemID string) ([]string, *logError) {
	medIDs := []string{}
	for link := range e.Map {
		if link.ItemID == itemID {
			medIDs = append(medIDs, link.ERx)
		}
	}

	if len(medIDs) == 0 {
		return nil, &logError{
			errMessage: fmt.Sprintf("error. no medIDs linked to itemID %s", itemID),
			logMessage: fmt.Sprintf("Error. No medIDs linked to itemID %s", itemID),
		}
	}

	return medIDs, nil
}

func (e *ERxItemIdLinks) GetAllItemIds() []string {
	itemIDsMap := map[string]struct{}{}
	itemIDs := []string{}

	for link := range e.Map {
		itemIDsMap[link.ItemID] = struct{}{}
	}

	for id := range itemIDsMap {
		itemIDs = append(itemIDs, id)
	}

	return itemIDs
}

func (e *ERxItemIdLinks) GetAssociatedItemIds(itemId string) []string {
	medIds, logErr := e.GetMedIds(itemId)
	if logErr != nil {
		return []string{itemId}
	}

	itemIdMap := map[string]struct{}{}
	for _, medId := range medIds {
		itemIds, _ := e.GetItemIds(medId)
		for _, id := range itemIds {
			itemIdMap[id] = struct{}{}
		}
	}

	itemIds := []string{}
	for id := range itemIdMap {
		itemIds = append(itemIds, id)
	}

	sort.Slice(itemIds, func(i, j int) bool {
		return itemIds[i] < itemIds[j]
	})

	return itemIds
}

func initERxItemIdLink() *ERxItemIdLinks {
	e := ERxItemIdLinks{}
	e.Map = map[ERxItemIdLink]struct{}{}
	return &e
}

func (p *Process) saveERxItemIdLinks() error {
	p.logger.LogInfo("Saving ERxItemIdLinks")

	file, err := os.OpenFile(filepath.Join(p.pathToData, ERxItemIdLinksFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error. Unable to open %s: %s", ERxItemIdLinksFileName, err.Error()))
		return err
	}
	defer file.Close()

	saveMap := map[string]ERxItemIdLink{}
	for link := range p.erxItemIdLinks.Map {
		saveMap[link.ERx] = link
	}
	saveLinks := ERxItemIdLinkSaveFormat{
		Map: saveMap,
	}

	data, err := json.Marshal(saveLinks)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error marshalling ERxItemIdLinks: %s", err.Error()))
		return err
	}

	_, err = file.Write(data)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Error writting marshalled ERxItemIdLinks to file %s: %s", ERxItemIdLinksFileName, err.Error()))
		return err
	}

	return nil

}

func (p *Process) loadERxItemIdLinks() error {
	p.logger.LogInfo("Loading ERxItemIdLinks")

	//-- Check if save file exists. Create file if not found.
	if _, err := os.Stat(filepath.Join(p.pathToData, ERxItemIdLinksFileName)); err != nil {
		if os.IsNotExist(err) {
			p.logger.LogInfo(fmt.Sprintf("%s not found. Attempting to create file", ERxItemIdLinksFileName))
			file, err := os.OpenFile(filepath.Join(p.pathToData, ERxItemIdLinksFileName), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				p.logger.LogError(fmt.Sprintf("Error. Unable to create ERxItemIdLinks save file: %s", err.Error()))
				return fmt.Errorf("error. unable to create ERxItemIdLinks save file: %s", err.Error())
			}
			_, err = file.WriteString("{}")
			if err != nil {
				p.logger.LogError(fmt.Sprintf("Error. Unable to write ERxItemIdLinks save file: %s", err.Error()))
				return fmt.Errorf("error. unable to write ERxItemIdLinks save file: %s", err.Error())
			}
			file.Close()
		}
	}

	//-- Load json data from save file
	data, err := os.ReadFile(filepath.Join(p.pathToData, ERxItemIdLinksFileName))
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err
	}

	//-- Unmarshal json into process memory
	saveLinks := ERxItemIdLinkSaveFormat{}
	err = json.Unmarshal(data, &saveLinks)
	if err != nil {
		p.logger.LogError(fmt.Sprintf("Load failed: %s", err.Error()))
		return err
	}

	for _, val := range saveLinks.Map {
		p.erxItemIdLinks.Map[val] = struct{}{}
	}

	p.logger.LogInfo("ERxItemIdLinks loaded successfully")
	p.state.ERxItemIdLinksSuccessful()
	return nil
}
