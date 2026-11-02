package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/cli"
)

func (p *Process) setupCommands() {
	p.cliConfig.AddCommand("hello", func([]cli.CommandArg) error {
		fmt.Println("Hello, World!")
		p.logger.LogInfo("Command hello executed")
		return nil
	})

	p.cliConfig.AddCommand("exit", func([]cli.CommandArg) error {
		p.logger.LogInfo("Command exit executed")
		fmt.Println("closing... goodbye!")
		p.exit()
		return nil
	})

	p.cliConfig.AddCommand("echo", func(args []cli.CommandArg) error {
		if len(args) == 0 {
			return fmt.Errorf("error. need phrase argument")
		}

		if len(args) > 1 {
			return fmt.Errorf("error. too many arguments")
		}

		fmt.Println(args[0].Val)
		return nil
	}, cli.CommandArg{
		Name: "phrase",
	})

	p.cliConfig.AddCommand("clear", func(args []cli.CommandArg) error {
		switch runtime.GOOS {
		case "windows":
			cmd := exec.Command("cmd", "/c", "cls")
			cmd.Stdout = os.Stdout
			cmd.Run()

		default:
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			cmd.Run()
		}

		return nil

	})

	p.cliConfig.AddCommand("save", func(args []cli.CommandArg) error {
		p.logger.LogInfo("save command executed")
		fmt.Println("saving data...")
		p.save()
		fmt.Println("save complete")
		return nil
	})

	p.cliConfig.AddCommand("add pyxis", func(args []cli.CommandArg) error {
		p.logger.LogInfo("Add pyxis command executed")
		pyxisName := ""

		for _, arg := range args {
			switch arg.Name {
			case "name":
				pyxisName = arg.Val

			}
		}

		err := p.addNewFacilityPyxis(pyxisName)
		if err != nil {
			return err
		}

		return nil

	}, cli.CommandArg{
		Name:     "name",
		Required: true,
	}, cli.CommandArg{
		Name:     "start_date",
		Required: true,
	})

	p.cliConfig.AddCommand("status", func(args []cli.CommandArg) error {
		p.logger.LogInfo("Status command executed")
		mode := p.state.Mode()
		p.logger.LogInfo(fmt.Sprintf("Current mode: %d", mode))

		fmt.Println(p.state.GetState())
		return nil
	})

	//---------------- ERx - ItemID Link Commands ------------------------//
	p.cliConfig.AddCommand("add ERxItemId link", func(args []cli.CommandArg) error {
		p.logger.LogInfo("add ERxItemId link command executed")

		erx := ""
		itemId := ""

		for _, arg := range args {
			switch arg.Name {
			case "erx":
				erx = arg.Val

			case "itemid":
				itemId = arg.Val
			}
		}

		//-- Check for blank args
		if erx == "" {
			p.logger.LogError("Command failed: erx cannot be blank")
			return fmt.Errorf("error. erx cannot be blank")
		}
		if itemId == "" {
			p.logger.LogError("Command failed: itemid cannot be blank")
			return fmt.Errorf("error. itemid cannot be blank")
		}

		//-- Check for letters in args
		_, err := strconv.Atoi(erx)
		if err != nil {
			p.logger.LogError("Command failed: erx can only contain numbers")
			return fmt.Errorf("error. erx can only contain numbers")
		}
		_, err = strconv.Atoi(itemId)
		if err != nil {
			p.logger.LogError("Command failed: itemid can only contain numbers")
			return fmt.Errorf("error. itemid can only contain numbers")
		}

		//-- Call method to add link
		logErr := p.erxItemIdLinks.Add(erx, itemId)
		if logErr != nil {
			p.logger.LogError(logErr.LogError())
			return logErr
		}
		p.logger.LogInfo(fmt.Sprintf("Link created. ItemId %s now links to ERx %s", itemId, erx))
		fmt.Println("link created")
		return nil

	}, cli.CommandArg{
		Name:     "erx",
		Required: true,
	}, cli.CommandArg{
		Name:     "itemid",
		Required: true,
	})

	p.cliConfig.AddCommand("remove ERxItemId link", func(args []cli.CommandArg) error {
		p.logger.LogInfo("remove ERxItemId Link command executed")

		erx := ""
		itemId := ""

		for _, arg := range args {
			switch arg.Name {
			case "erx":
				erx = arg.Val

			case "itemId":
				itemId = arg.Val
			}
		}

		if erx == "" {
			p.logger.LogError("Command failed: erx cannot be blank")
			return fmt.Errorf("error. erx cannot be blank")
		}

		if itemId == "" {
			p.logger.LogError("Command failed: itemId cannot be blank")
			return fmt.Errorf("error. itemId cannot be blank")
		}

		logErr := p.erxItemIdLinks.Remove(erx, itemId)
		if logErr != nil {
			p.logger.LogError("Command failed: " + logErr.logMessage)
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("erx %s link removed from itemID %s", erx, itemId))
		fmt.Printf("erx %s link removed from itemID %s\n", erx, itemId)
		return nil

	}, cli.CommandArg{
		Name:     "erx",
		Required: true,
	}, cli.CommandArg{
		Name:     "itemId",
		Required: true,
	})

	p.cliConfig.AddCommand("list all itemID associations", func(args []cli.CommandArg) error {
		p.logger.LogInfo("list all itemID associations command executed")

		itemId := ""
		for _, arg := range args {
			switch arg.Name {
			case "itemId":
				itemId = arg.Val
			}
		}

		if itemId == "" {
			p.logger.LogError("Command failed. ItemId cannot be blank")
			return fmt.Errorf("error. itemId cannot be blank")
		}

		printfln("ItemID %s %s is directly linked to the following MedIDs:", itemId, quickDisplayName(p.itemIDs.DisplayName, itemId))
		medIDs, _ := p.erxItemIdLinks.GetMedIds(itemId)
		for _, id := range medIDs {
			printfln("   %s[%s]", quickDisplayName(p.erxs.DisplayName, id), id)
		}
		fmt.Println()

		printfln("ItemID %s %s is associated with the following ItemIDs:", itemId, quickDisplayName(p.itemIDs.DisplayName, itemId))

		for _, id := range p.erxItemIdLinks.GetAssociatedItemIds(itemId) {
			printfln("   %s %s", id, quickDisplayName(p.itemIDs.DisplayName, id))
		}
		fmt.Println()

		printfln("ItemED %s %s is associated with the following MedIDs:", itemId, quickDisplayName(p.itemIDs.DisplayName, itemId))

		medIdMap := map[string]struct{}{}
		for _, id := range p.erxItemIdLinks.GetAssociatedItemIds(itemId) {
			medIds, _ := p.erxItemIdLinks.GetMedIds(id)
			for _, medId := range medIds {
				medIdMap[medId] = struct{}{}
			}
		}
		for id := range medIdMap {
			printfln("   %s[%s]", quickDisplayName(p.erxs.DisplayName, id), id)
		}

		fmt.Println()
		return nil

	}, cli.CommandArg{
		Name:     "itemId",
		Required: true,
	})

	//-------------------- Department Coverage Commands -------------------------//
	p.cliConfig.AddCommand("add department coverage", func(args []cli.CommandArg) error {
		p.logger.LogInfo("add department coverage command executed")

		deptID := ""
		deptName := ""

		for _, arg := range args {
			switch arg.Name {
			case "deptID":
				deptID = arg.Val

			case "dept_name":
				deptName = arg.Val
			}
		}

		if deptID == "" {
			err := fmt.Errorf("error. deptID cannot be blank")
			p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
			return err
		}

		if !isNumeric(deptID) {
			err := fmt.Errorf("error. deptID must contain only numbers")
			p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
			return err
		}

		if deptName == "" {
			err := fmt.Errorf("error. dept_name cannot be blank")
			p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
			return err
		}

		err := p.departmentCoverage.addNewFacilityDepartment(deptID, deptName)
		if err != nil {
			p.logger.LogError("Command failed: " + err.Error())
			return err
		}

		p.logger.LogInfo(fmt.Sprintf("Department %s [%s] added", deptName, deptID))
		return nil

	}, cli.CommandArg{
		Name:     "deptID",
		Required: true,
	}, cli.CommandArg{
		Name: "dept_name",
	})

	p.cliConfig.AddCommand("remove department coverage", func(args []cli.CommandArg) error {
		p.logger.LogInfo("remove department coverage command executed")

		deptID := ""

		for _, arg := range args {
			switch arg.Name {
			case "deptID":
				deptID = arg.Val
			}
		}

		if deptID == "" {
			p.logger.LogError("Command failed: deptID cannot be blank")
			return fmt.Errorf("error. deptID cannot be blank")
		}

		deptName, err := p.departmentCoverage.GetDeptName(deptID)
		if err != nil {
			p.logger.LogError("Command Failed: " + err.Error())
			return err
		}

		err = p.departmentCoverage.removeFacilityDepartment(deptID)
		if err != nil {
			p.logger.LogError("Command Failed: " + err.Error())
			return err
		}

		p.logger.LogInfo(fmt.Sprintf("%s [%s] removed from %s as a covered department", deptName, deptID))
		fmt.Printf("%s [%s] removed from %s as a covered department\n", deptName, deptID)
		return nil

	}, cli.CommandArg{
		Name:     "deptID",
		Required: true,
	})

	p.cliConfig.AddCommand("list department coverage", func(args []cli.CommandArg) error {
		p.logger.LogInfo("list department coverage command executed")

		depts := p.departmentCoverage.GetCoveredDepartments()

		fmt.Println("Departments Covered:")
		for _, dept := range depts {
			fmt.Printf("%s [%s]\n", dept.Name, dept.ID)
		}

		return nil

	})

	//------------------ ERX Commands ----------------------//
	p.cliConfig.AddCommand("list erxs", func(args []cli.CommandArg) error {
		p.logger.LogInfo("list ERXs command executed")

		erxs := p.erxs.GetAll()
		for _, erx := range erxs {
			fmt.Printf("%s [%s]\n", erx.DisplayName, erx.MedID)
		}

		return nil
	})

	p.cliConfig.AddCommand("add erx", func(args []cli.CommandArg) error {
		p.logger.LogInfo("add erx command executed")

		erx := ""
		name := ""

		for _, arg := range args {
			switch arg.Name {
			case "erx":
				erx = arg.Val

			case "name":
				name = arg.Val
			}
		}

		if erx == "" {
			p.logger.LogError("Command Failed: erx cannot be blank")
			return fmt.Errorf("error. erx cannot be blank")
		}
		if name == "" {
			p.logger.LogError("Command failed: name cannot be blank")
			return fmt.Errorf("error name cannot be blank")
		}

		logErr := p.erxs.Add(erx, name)
		if logErr != nil {
			p.logger.LogError("Command failed: " + logErr.logMessage)
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("%s [%s] added to ERXs", name, erx))
		fmt.Printf("%s [%s] added to ERXs\n", name, erx)
		return nil

	}, cli.CommandArg{
		Name:     "erx",
		Required: true,
	}, cli.CommandArg{
		Name:     "name",
		Required: true,
	})

	p.cliConfig.AddCommand("remove erx", func(args []cli.CommandArg) error {
		p.logger.LogInfo("remove erx command executed")

		erx := ""
		for _, arg := range args {
			switch arg.Name {
			case "erx":
				erx = arg.Val
			}
		}

		if erx == "" {
			p.logger.LogError("Command failed: erx cannot be blank")
			return fmt.Errorf("error. erx cannot be blank")
		}

		logErr := p.erxs.Remove(erx)
		if logErr != nil {
			p.logger.LogError("Command failed: " + logErr.logMessage)
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("erx %s removed", erx))
		fmt.Printf("erx %s removed\n", erx)
		return nil

	}, cli.CommandArg{
		Name:     "erx",
		Required: true,
	})

	//----------------------ItemID Commands -------------------------//
	p.cliConfig.AddCommand("add itemID", func(args []cli.CommandArg) error {
		p.logger.LogInfo("add itemID command executed")

		itemID := ""
		name := ""

		for _, arg := range args {
			switch arg.Name {
			case "itemID":
				itemID = arg.Val

			case "name":
				name = arg.Val
			}
		}

		if itemID == "" {
			p.logger.LogError("Command failed: itemID cannot be blank")
			return fmt.Errorf("error. itemID cannot be blank")
		}

		if name == "" {
			p.logger.LogError("Command failed: name cannot be blank")
			return fmt.Errorf("error. name cannot be blank")
		}

		logErr := p.itemIDs.Add(itemID, name)
		if logErr != nil {
			p.logger.LogError("Command failed: " + logErr.logMessage)
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("ItemID %s %s added", itemID, name))
		fmt.Printf("itemID %s %s added\n", itemID, name)
		return nil

	}, cli.CommandArg{
		Name:     "itemID",
		Required: true,
	}, cli.CommandArg{
		Name:     "name",
		Required: true,
	})

	p.cliConfig.AddCommand("remove itemID", func(args []cli.CommandArg) error {
		p.logger.LogInfo("remove itemID command executed")

		itemID := ""

		for _, arg := range args {
			switch arg.Name {
			case "itemID":
				itemID = arg.Val
			}
		}

		if itemID == "" {
			p.logger.LogError("Command failed: itemID cannot be blank")
			return fmt.Errorf("error. itemID cannot be blank")
		}

		logErr := p.itemIDs.Remove(itemID)
		if logErr != nil {
			p.logger.LogError("Command failed: " + logErr.logMessage)
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("ItemID %s removed", itemID))
		fmt.Printf("ItemID %s removed\n", itemID)
		return nil

	}, cli.CommandArg{
		Name:     "itemID",
		Required: true,
	})

	p.cliConfig.AddCommand("list itemIDs", func(args []cli.CommandArg) error {
		p.logger.LogInfo("list itemIDs command executed")

		itemIDs := p.itemIDs.GetAll()
		for _, itemID := range itemIDs {
			fmt.Printf("%s %s\n", itemID.ID, itemID.DisplayName)
		}

		return nil
	})

	//-------------------- Control Event Trail Commands ------------------------//
	p.cliConfig.AddCommand("generate control trail", func(args []cli.CommandArg) error {
		p.logger.LogInfo("generate control trail command executed")

		type ReportRowValue struct {
			Key         string
			Type        string
			DateTime    string
			UserID      string
			UserName    string
			DisplayName string
			Amount      string
			Units       string
			MRN         string
			WitPtName   string
		}

		type RowNames struct {
			KeyName         string
			TypeName        string
			DateTimeName    string
			UserIDName      string
			UserNameName    string
			DisplayNameName string
			AmountName      string
			UnitsName       string
			MrnName         string
			WitPtName       string
		}

		eventRowNames := RowNames{
			KeyName:         "Item Transaction Key",
			TypeName:        "Transaction Type",
			DateTimeName:    "Tx Date Time",
			UserIDName:      "User ID",
			UserNameName:    "User Name",
			DisplayNameName: "Display Name",
			AmountName:      "Amount",
			MrnName:         "MRN",
			WitPtName:       "Witness",
		}

		actionRowNames := RowNames{
			KeyName:         "Order Number",
			TypeName:        "MAR Action",
			DateTimeName:    "Saved Time",
			UserIDName:      "User ID",
			UserNameName:    "User Name",
			DisplayNameName: "Display Name",
			AmountName:      "Dose",
			MrnName:         "MRN",
			WitPtName:       "Patient Name",
		}

		correctionRowNames := RowNames{
			KeyName:         "Key",
			TypeName:        "Write Up File",
			DateTimeName:    "Correction Date",
			UserIDName:      "User ID",
			UserNameName:    "User Name",
			DisplayNameName: "Display Name",
			AmountName:      "Correction Amount",
			MrnName:         "MRN",
			WitPtName:       "Patient Name",
		}

		pyxis := ""
		for _, arg := range args {
			switch arg.Name {
			case "pyxis":
				pyxis = arg.Val
			}
		}

		if pyxis == "" {
			p.logger.LogError("Command failed. Pyxis cannot be blank")
			return fmt.Errorf("error. pyxis cannot be blank")
		}

		for _, pyxisEventLog := range p.PyxisEventLogs {
			if pyxisEventLog.PyxisName == pyxis {
				controlTrailSlices := pyxisEventLog.controlEventLog.GenerateTrailSlices()

				report := [][]string{}

				for _, controlTrailSlice := range controlTrailSlices {
					for _, controlTrail := range controlTrailSlice {
						batch := make([][]string, 10)
						for i := range batch {
							batch[i] = make([]string, len(controlTrail.Trail)*2)
						}
						for x, event := range controlTrail.Trail {
							switch event.Type {
							case pyxisEvent:
								batch[0][x*2] = eventRowNames.KeyName
								batch[0][(x*2)+1] = event.PyxisEvent.ItemTransactionKey.String()
								batch[1][x*2] = eventRowNames.TypeName
								batch[1][(x*2)+1] = event.PyxisEvent.TransactionType
								batch[2][x*2] = eventRowNames.DateTimeName
								batch[2][(x*2)+1] = event.PyxisEvent.TxDateTime.Format("2006-01-02 1504")
								batch[3][x*2] = eventRowNames.UserIDName
								batch[3][(x*2)+1] = event.PyxisEvent.UserID
								batch[4][x*2] = eventRowNames.UserNameName
								batch[4][(x*2)+1] = event.PyxisEvent.UserName
								batch[5][x*2] = eventRowNames.DisplayNameName
								batch[5][(x*2)+1] = event.PyxisEvent.MedDisplayName
								batch[6][x*2] = eventRowNames.AmountName
								batch[6][(x*2)+1] = fmt.Sprintf("%s %s",
									strconv.FormatFloat(event.PyxisEvent.AmountReferenced, 'f', -1, 64),
									event.PyxisEvent.AmountReferencedUnits)
								batch[7][x*2] = eventRowNames.MrnName
								batch[7][(x*2)+1] = event.PyxisEvent.MRN
								batch[8][x*2] = eventRowNames.WitPtName
								batch[8][(x*2)+1] = event.PyxisEvent.WitnessName

							case marAction:
								batch[0][x*2] = actionRowNames.KeyName
								batch[0][(x*2)+1] = event.MarAction.OrderNumber
								batch[1][x*2] = actionRowNames.TypeName
								batch[1][(x*2)+1] = event.MarAction.MarAction
								batch[2][x*2] = actionRowNames.DateTimeName
								batch[2][(x*2)+1] = event.MarAction.SavedTime.Format("2006-01-02 1504")
								batch[3][x*2] = actionRowNames.UserIDName
								batch[3][(x*2)+1] = event.MarAction.UserID
								batch[4][x*2] = actionRowNames.UserNameName
								batch[4][(x*2)+1] = event.MarAction.UserName
								batch[5][x*2] = actionRowNames.DisplayNameName
								batch[5][(x*2)+1] = event.MarAction.DisplayName
								batch[6][x*2] = actionRowNames.AmountName
								batch[6][(x*2)+1] = fmt.Sprintf("%s %s",
									strconv.FormatFloat(event.MarAction.Dose, 'f', -1, 64),
									event.MarAction.DoseUnitDescription)
								batch[7][x*2] = actionRowNames.MrnName
								batch[7][(x*2)+1] = event.MarAction.MRN
								batch[8][x*2] = actionRowNames.WitPtName
								batch[8][(x*2)+1] = event.MarAction.PtName

							case correctionEvent:
								batch[0][x*2] = correctionRowNames.KeyName
								batch[0][(x*2)+1] = event.CorrectionEvent.Id
								batch[1][x*2] = correctionRowNames.TypeName
								batch[1][(x*2)+1] = p.correctionEventLinks.WriteUpFile(event.CorrectionEvent.Id)
								batch[2][x*2] = correctionRowNames.DateTimeName
								batch[2][(x*2)+1] = event.CorrectionEvent.CorrectionDate.Format("2006-01-02")
								batch[3][x*2] = correctionRowNames.UserIDName
								batch[3][(x*2)+1] = event.CorrectionEvent.UserID
								batch[4][x*2] = correctionRowNames.UserNameName
								batch[4][(x*2)+1] = event.CorrectionEvent.UserName
								batch[5][x*2] = correctionRowNames.DisplayNameName
								batch[5][(x*2)+1] = event.CorrectionEvent.DisplayName
								batch[6][x*2] = correctionRowNames.AmountName
								batch[6][(x*2)+1] = fmt.Sprintf("%s %s",
									strconv.FormatFloat(event.CorrectionEvent.Amount, 'f', -1, 64),
									event.CorrectionEvent.Units)
								batch[7][x*2] = correctionRowNames.MrnName
								batch[7][(x*2)+1] = event.CorrectionEvent.MRN
								batch[8][x*2] = correctionRowNames.WitPtName
								batch[8][(x*2)+1] = event.CorrectionEvent.PtName

							}
						}
						report = append(report, batch...)
					}
				}

				file, err := os.OpenFile(filepath.Join(p.pathToOut, pyxis+"_ControlEventTrails"+".csv"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					p.logger.LogError(fmt.Sprintf("Command failed: Error opening %s_ControlEventTrails.csv : %s", pyxis, err.Error()))
					return err
				}
				defer file.Close()

				writer := csv.NewWriter(file)
				err = writer.WriteAll(report)
				if err != nil {
					p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
					fmt.Println(err.Error())
				}

				return nil
			}
		}

		p.logger.LogError(fmt.Sprintf("Command failed. %s pyxis not found", pyxis))
		return fmt.Errorf("error. %s pyxis not found", pyxis)

	}, cli.CommandArg{
		Name:     "pyxis",
		Required: true,
	})

	p.cliConfig.AddCommand("list unmatched events", func(args []cli.CommandArg) error {
		p.logger.LogInfo("list unmatched events command executed")

		pyxis := ""

		for _, arg := range args {
			switch arg.Name {
			case "pyxis":
				pyxis = arg.Val
			}
		}

		if pyxis == "" {
			p.logger.LogError("Command failed. Pyxis cannot be blank")
			return fmt.Errorf("error. pyxis cannot be blank")
		}

		index := 0
		found := false
		for i, p := range p.PyxisEventLogs {
			if p.PyxisName == pyxis {
				index = i
				found = true
				break
			}
		}
		if !found {
			p.logger.LogError(fmt.Sprintf("Command failed. %s Pyxis not found", pyxis))
			return fmt.Errorf("error. %s pyxis not found", pyxis)
		}

		p.PyxisEventLogs[index].controlEventLog.SortUnmatchedEvents()

		for _, unmatchedEvent := range p.PyxisEventLogs[index].controlEventLog.UnmatchedEvents {
			fmt.Printf("Key: %s\nType: %s\nDateTime: %s\nItemID: %s\nUserName: %s (%s)\nDisplayName: %s\nAmount: %s\nMRN: %s\nWitness: %s\n",
				unmatchedEvent.ItemTransactionKey.String(),
				unmatchedEvent.TransactionType,
				unmatchedEvent.TxDateTime.Format("2006-01-02 1504"),
				unmatchedEvent.ItemID,
				unmatchedEvent.UserName,
				unmatchedEvent.UserID,
				unmatchedEvent.MedDisplayName,
				strconv.FormatFloat(unmatchedEvent.AmountReferenced, 'f', -1, 64)+" "+unmatchedEvent.AmountReferencedUnits,
				unmatchedEvent.MRN,
				unmatchedEvent.WitnessName)
			fmt.Println()
		}

		return nil

	}, cli.CommandArg{
		Name:     "pyxis",
		Required: true,
	})

	//-------------------------- Event Action Selection Commands ------------------------------//
	p.cliConfig.AddCommand("select unmatched event", func(args []cli.CommandArg) error {
		p.logger.LogInfo("select unmatched event command executed")

		pyxis := ""
		id := ""

		for _, arg := range args {
			switch arg.Name {
			case "pyxis":
				pyxis = arg.Val

			case "id":
				id = arg.Val
			}
		}

		if pyxis == "" {
			p.logger.LogError("Command failed. Pyxis cannot be blank")
			return fmt.Errorf("error. pyxis cannot be blank")
		}

		if id == "" {
			p.logger.LogError("Command failed. Id cannot be blank")
			return fmt.Errorf("error. id cannot be blank")
		}

		iLog := 0
		found := false
		for i, log := range p.PyxisEventLogs {
			if log.PyxisName == pyxis {
				found = true
				iLog = i
				break
			}
		}
		if !found {
			p.logger.LogError(fmt.Sprintf("Command failed. %s Pyxis not found", pyxis))
			return fmt.Errorf("error. %s pyxis not found", pyxis)
		}

		index := 0
		found = false
		for i, unmatchedEvent := range p.PyxisEventLogs[iLog].controlEventLog.UnmatchedEvents {
			if unmatchedEvent.ItemTransactionKey.String() == id {
				found = true
				index = i
			}
		}
		if !found {
			p.logger.LogError(fmt.Sprintf("Command failed. Unmatched event %s not found", id))
			return fmt.Errorf("error. unmatched event %s not found", id)
		}

		logErr := p.selectedEventActions.SelectUnmatchedEvent(p.PyxisEventLogs[iLog].controlEventLog, index)
		if logErr != nil {
			p.logger.LogError(fmt.Sprintf("Command failed: %s", logErr.logMessage))
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("Unmatched event %s added to selection", id))
		printfln("unmatched event %s added to selection", id)
		return nil

	}, cli.CommandArg{
		Name:     "pyxis",
		Required: true,
	}, cli.CommandArg{
		Name:     "id",
		Required: true,
	})

	p.cliConfig.AddCommand("select mar action", func(args []cli.CommandArg) error {
		p.logger.LogInfo("select mar action command executed")

		if !p.state.DbConnectionOkay() {
			p.logger.LogError("Command failed. Database connection required for this command")
			return fmt.Errorf("error. database connection required for this command")
		}

		orderNumber := ""
		for _, arg := range args {
			switch arg.Name {
			case "orderNumber":
				orderNumber = arg.Val
			}
		}

		if orderNumber == "" {
			p.logger.LogError("Command failed. orderNumber cannot be blank")
			return fmt.Errorf("error. orderNumber cannot be blank")
		}

		if !isNumeric(orderNumber) {
			p.logger.LogError("Command failed. order number must only contain numbers")
			return fmt.Errorf("error. orderNumber must only contain numbers")
		}

		responseActions, err := getMarActionsByOrderNumber(p, orderNumber)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
			return err
		}

		actions := p.parseMarActions(responseActions)

		event := EventTrailItem{
			Type: marAction,
		}
		switch len(actions) {
		case 0:
			p.logger.LogError(fmt.Sprintf("Command failed. No MAR actions found with order number %s", orderNumber))
			return fmt.Errorf("error. no mar actions found with order number %s", orderNumber)

		case 1:
			event.MarAction = actions[0]

		default:
			fmt.Println("Select which event to select:")
			selectActionScanner := bufio.NewScanner(os.Stdin)
			input := ""

			for i, action := range actions {
				printfln("%d. MAR Action: %s", i+1, action.MarAction)
				printfln("   Date Time: %s", action.SavedTime.Format("2006-01-02 1504"))
				printfln("   Display Name: %s", action.DisplayName)
				printfln("   Dose: %s %s", strconv.FormatFloat(action.Dose, 'f', -1, 64), action.DoseUnitDescription)
				printfln("   User: %s (%s)", action.UserName, action.UserID)
				fmt.Println()
			}

			selectActionScanner.Scan()
			input = selectActionScanner.Text()
			index, err := strconv.Atoi(input)
			if err != nil {
				p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
				return err
			}
			index--

			if index >= len(actions) {
				p.logger.LogError(fmt.Sprintf("Command failed. Invalid input: %s", input))
				return fmt.Errorf("error. invalid input %s", input)
			}

			event.MarAction = actions[index]
		}

		logErr := p.selectedEventActions.Add(event)
		if logErr != nil {
			p.logger.LogError(fmt.Sprintf("Command failed: %s", logErr.logMessage))
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("MAR Action with order number %s added to selection", orderNumber))

		fmt.Println("MAR Action added to selection: ")
		printfln("MAR Action: %s", event.MarAction.MarAction)
		printfln("Date Time: %s", event.MarAction.SavedTime.Format("2006-01-02 1504"))
		printfln("Display Name: %s", event.MarAction.DisplayName)
		printfln("Dose: %s %s", strconv.FormatFloat(event.MarAction.Dose, 'f', -1, 64), event.MarAction.DoseUnitDescription)
		printfln("User: %s (%s)", event.MarAction.UserName, event.MarAction.UserID)
		fmt.Println()

		return nil

	}, cli.CommandArg{
		Name:     "orderNumber",
		Required: true,
	})

	p.cliConfig.AddCommand("select pyxis event", func(args []cli.CommandArg) error {
		p.logger.LogInfo("select pyxis event command executed")

		pyxis := ""
		id := ""

		for _, arg := range args {
			switch arg.Name {
			case "pyxis":
				pyxis = arg.Val

			case "id":
				id = arg.Val
			}
		}

		if pyxis == "" {
			p.logger.LogError("Command failed. pyxis cannot be blank")
			return fmt.Errorf("error. pyxis cannot be blank")
		}

		if id == "" {
			p.logger.LogError("Command failed. id cannot be blank")
			return fmt.Errorf("error. id cannot be blank")
		}

		err := p.loadPyxisEventlog(pyxis)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
			return err
		}

		logIndex := 0
		found := false
		for i, log := range p.PyxisEventLogs {
			if log.PyxisName == pyxis {
				found = true
				logIndex = i
			}
		}
		if !found {
			p.logger.LogError(fmt.Sprintf("Command failed. %s Pyxis not found", pyxis))
			return fmt.Errorf("error. %s pyxis not found", pyxis)
		}

		selectedEvent := PyxisEvent{}
		found = false
		for _, event := range p.PyxisEventLogs[logIndex].Log {
			if event.ItemTransactionKey.String() == id {
				found = true
				selectedEvent = event
			}
		}
		if !found {
			p.logger.LogError(fmt.Sprintf("Command failed. Pyxis event %s not found", id))
			return fmt.Errorf("error. pyxis event %s not found", id)
		}

		err = p.saveAndUnloadPyxisEventLogs()
		if err != nil {
			p.logger.LogError(err.Error())
		}

		selectedItem := EventTrailItem{
			Type:       pyxisEvent,
			PyxisEvent: selectedEvent,
		}

		logErr := p.selectedEventActions.Add(selectedItem)
		if logErr != nil {
			p.logger.LogError(fmt.Sprintf("Command failed: %s", logErr.logMessage))
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("Pyxis event %s from %s added to selected event actions", id, pyxis))
		printfln("pyxis event %s from %s added to selected event actions", id, pyxis)
		return nil

	}, cli.CommandArg{
		Name:     "pyxis",
		Required: true,
	}, cli.CommandArg{
		Name:     "id",
		Required: true,
	})

	p.cliConfig.AddCommand("get and select pyxis event", func(args []cli.CommandArg) error {
		p.logger.LogInfo("get and select pyxis event command executed")

		if !p.state.DbConnectionOkay() {
			p.logger.LogError("Command failed. Database connection required for this command")
			return fmt.Errorf("error. database connection required for this command")
		}

		id := ""
		for _, arg := range args {
			switch arg.Name {
			case "id":
				id = arg.Val
			}
		}

		if id == "" {
			p.logger.LogError("Command failed. id cannot be blank")
			return fmt.Errorf("error. id cannot be blank")
		}

		eventResponse, err := getPyxisEventById(p, id)
		if err != nil {
			p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
			return err
		}

		event := ParseEvent(eventResponse)

		item := EventTrailItem{
			Type:       pyxisEvent,
			PyxisEvent: event,
		}

		logErr := p.selectedEventActions.Add(item)
		if logErr != nil {
			p.logger.LogError(fmt.Sprintf("Command failed: %s", logErr.logMessage))
			return logErr
		}

		p.logger.LogInfo(fmt.Sprintf("Pyxis event %s added to selection", id))
		printfln("pyxis event %s added to selection", id)

		return nil

	}, cli.CommandArg{
		Name:     "id",
		Required: true,
	})

	p.cliConfig.AddCommand("list selected items", func(args []cli.CommandArg) error {
		p.logger.LogInfo("list selected items command executed")

		items := []EventTrailItem{}
		for item := range p.selectedEventActions.Map {
			items = append(items, item)
		}

		sortEventTrailItems(items)
		printEventTrailItems(items)

		return nil

	})

	p.cliConfig.AddCommand("remove selected item", func(args []cli.CommandArg) error {
		p.logger.LogInfo("remove selected item command executed")

		id := ""
		for _, arg := range args {
			switch arg.Name {
			case "id":
				id = arg.Val
			}
		}

		if id == "" {
			p.logger.LogError("Command failed. id cannot be blank")
			return fmt.Errorf("error. id cannot be blank")
		}

		found := false
		itemToRemove := EventTrailItem{}
		for item := range p.selectedEventActions.Map {
			switch item.Type {
			case pyxisEvent:
				if item.PyxisEvent.ItemTransactionKey.String() == id {
					found = true
					itemToRemove = item
				}

			case marAction:
				if item.MarAction.OrderNumber == id {
					found = true
					itemToRemove = item
				}

			case correctionEvent:
				if item.CorrectionEvent.Id == id {
					found = true
					itemToRemove = item
				}
			}
		}
		if !found {
			p.logger.LogError(fmt.Sprintf("Command failed. Id %s not found in selected items", id))
			return fmt.Errorf("error. id %s not found in selected items", id)
		}

		delete(p.selectedEventActions.Map, itemToRemove)

		switch itemToRemove.Type {
		case pyxisEvent:
			printfln("Pyxis event %s removed from selected items", itemToRemove.PyxisEvent.ItemTransactionKey.String())

		case marAction:
			printfln("MAR action taken on %s at %s with order number %s removed from selected items",
				itemToRemove.MarAction.SavedTime.Format("2006-01-02"),
				itemToRemove.MarAction.SavedTime.Format("1504"),
				itemToRemove.MarAction.OrderNumber)

		case correctionEvent:
			printfln("Correction action %s removed from selected items", id)
		}

		return nil

	}, cli.CommandArg{
		Name:     "id",
		Required: true,
	})

	p.cliConfig.AddCommand("break event trail and select", func(args []cli.CommandArg) error {
		p.logger.LogInfo("break event trail and select command executed")

		startLen := len(p.selectedEventActions.Map)
		pyxis := ""
		id := ""

		for _, arg := range args {
			switch arg.Name {
			case "pyxis":
				pyxis = arg.Val

			case "id":
				id = arg.Val
			}
		}

		if pyxis == "" {
			p.logger.LogError("Command failed. pyxis cannot be blank")
			return fmt.Errorf("error. pyxis cannot be blank")
		}

		if id == "" {
			p.logger.LogError("Command failed. id cannot be blank")
			return fmt.Errorf("error. id cannot be blank")
		}

		logIndex := 0
		found := false
		for i, log := range p.PyxisEventLogs {
			if log.PyxisName == pyxis {
				found = true
				logIndex = i
			}
		}
		if !found {
			p.logger.LogError(fmt.Sprintf("Command failed. %s Pyxis not found", pyxis))
			return fmt.Errorf("error. %s pyxis not found", pyxis)
		}

		found = false
		for i := range p.PyxisEventLogs[logIndex].controlEventLog.Log {
			for j := range p.PyxisEventLogs[logIndex].controlEventLog.Log[i].EventTrails {
				for _, item := range p.PyxisEventLogs[logIndex].controlEventLog.Log[i].EventTrails[j].Trail {
					switch item.Type {
					case pyxisEvent:
						if item.PyxisEvent.ItemTransactionKey.String() == id {
							found = true
							logErr := p.selectedEventActions.RemoveTrailAndSelect(p.PyxisEventLogs[logIndex].controlEventLog, &p.PyxisEventLogs[logIndex].controlEventLog.Log[i], j)
							if logErr != nil {
								p.logger.LogError(fmt.Sprintf("Command failed: %s", logErr.logMessage))
								return logErr
							}
						}

					case marAction:
						//
					}
					if found {
						break
					}
				}
				if found {
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			p.logger.LogError(fmt.Sprintf("Command failed: %s not found in %s control event log", id, pyxis))
			return fmt.Errorf("error. %s not found in %s control event log", id, pyxis)
		}

		lenChange := len(p.selectedEventActions.Map) - startLen
		p.logger.LogInfo(fmt.Sprintf("%d event actions added to selection", lenChange))
		printfln("%d event actions added to selection", lenChange)

		return nil

	}, cli.CommandArg{
		Name:     "pyxis",
		Required: true,
	}, cli.CommandArg{
		Name:     "id",
		Required: true,
	})

	//----------------------- Manual Control Matching Commands ------------------------//

	p.cliConfig.AddCommand("match selected event actions", func(args []cli.CommandArg) error {
		p.logger.LogInfo("match selected event actions command executed")

		pyxis := ""
		mod := ""

		for _, arg := range args {
			switch arg.Name {
			case "pyxis":
				pyxis = arg.Val

			case "modifier":
				mod = arg.Val
			}
		}

		if pyxis == "" {
			p.logger.LogError("Command failed. Pyxis cannot be blank")
			return fmt.Errorf("error. pyxis cannot be blank")
		}

		if mod == "-c" { //-- If correction event to include
			newCorrectionEventScanner := bufio.NewScanner(os.Stdin)

			prompts := []string{
				"Event Date: ",      //-- 0
				"Correction Date: ", //-- 1
				"User ID: ",         //-- 2
				"User Name: ",       //-- 3
				"Item ID: ",         //-- 4
				"Display Name: ",    //-- 5
				"Amount: ",          //-- 6
				"Units: ",           //-- 7
				"MRN: ",             //-- 8
				"Patient Name: ",    //-- 9
				"BeSafe: ",          //-- 10
			}

			inputs := []string{}

			for i := 0; i < len(prompts); i++ {
				fmt.Print(prompts[i])
				newCorrectionEventScanner.Scan()
				input := newCorrectionEventScanner.Text()
				inputs = append(inputs, input)
			}

			eventDate, err := parseDate(inputs[0])
			if err != nil {
				p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
				return err
			}
			eventDate = timeStartDay(eventDate)

			correctionDate, err := parseDate(inputs[1])
			if err != nil {
				p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
				return err
			}

			itemID := inputs[4]
			_, logErr := p.itemIDs.DisplayName(itemID)
			if logErr != nil {
				p.logger.LogError(fmt.Sprintf("Command failed: itemID %s not found", itemID))
				return err
			}

			amount, err := strconv.ParseFloat(inputs[6], 64)
			if err != nil {
				p.logger.LogError(fmt.Sprintf("Command failed: %s", err.Error()))
				return err
			}

			dateID, index := p.correctionEventLinks.GetNewLink(p.pathToData, correctionDate)
			link := p.correctionEventLinks.Map[dateID][index]

			newCorrectionEvent := CorrectionEvent{
				Id:             link.Id,
				EventDate:      eventDate,
				CorrectionDate: correctionDate,
				UserID:         inputs[2],
				UserName:       inputs[3],
				ItemId:         itemID,
				DisplayName:    inputs[5],
				Amount:         amount,
				Units:          inputs[7],
				MRN:            inputs[8],
				PtName:         inputs[9],
				BeSafe:         inputs[10],
			}

			logErr = p.correctionEventLinks.AddAndLink(p, pyxis, eventDate, inputs[8], itemID, dateID, index, newCorrectionEvent)
			if logErr != nil {
				p.logger.LogError(fmt.Sprintf("Command failed: %s", logErr.logMessage))
				return logErr
			}

			err = openInDefaultEditor(link.WriteUpFile)
			if err != nil {
				p.logger.LogError("Error. Launching write up file failed")
				return err
			}

			//-- Remove selected event actions after successfully matching items
			p.selectedEventActions.Map = map[EventTrailItem]struct{}{}

			return nil

		} else { //-- If no correction event to include
			if pyxis == "" {
				p.logger.LogError("Command failed. Pyxis cannot be blank")
				return fmt.Errorf("error. pyxis cannot be blank")
			}

			logIndex := 0
			found := false
			for i, log := range p.PyxisEventLogs {
				if log.PyxisName == pyxis {
					found = true
					logIndex = i
					break
				}
			}
			if !found {
				p.logger.LogError(fmt.Sprintf("Command failed. %s Pyxis not found", pyxis))
				return fmt.Errorf("error. %s pyxis not found", pyxis)
			}

			events := []EventTrailItem{}
			for event := range p.selectedEventActions.Map {
				events = append(events, event)
			}

			if len(events) == 0 {
				p.logger.LogError("Command failed. no selected events")
				return fmt.Errorf("error. no selected events")
			}

			sortEventTrailItems(events)

			found = false
			index := 0
			for i, event := range events {
				if event.Type == correctionEvent {
					found = true
					index = i
					break
				}
			}
			if !found {
				for i, event := range events {
					if event.Type == pyxisEvent {
						found = true
						index = i
						break
					}
				}
				if !found {
					p.logger.LogError("Command failed. Selections must include at least one Pyxis event if a Correction event is not present")
					return fmt.Errorf("error. selections must include at least one pyxis event if a correction event is not present")
				}

			}

			mrn := ""
			itemId := ""
			date := time.Time{}

			switch events[index].Type {
			case correctionEvent:
				mrn = events[index].CorrectionEvent.MRN
				itemId = events[index].CorrectionEvent.ItemId
				date = timeStartDay(events[index].CorrectionEvent.EventDate)

			case pyxisEvent:
				mrn = events[index].PyxisEvent.MRN
				itemId = events[index].PyxisEvent.ItemID
				date = timeStartDay(events[index].PyxisEvent.TxDateTime)
			}

			logErr := p.PyxisEventLogs[logIndex].controlEventLog.LinkEventActions(mrn, itemId, date, events...)
			if logErr != nil {
				p.logger.LogError(fmt.Sprintf("Command failed: %s", logErr.logMessage))
				return logErr
			}

			//-- Remove selected event actions after successfully matching items
			p.selectedEventActions.Map = map[EventTrailItem]struct{}{}

			p.logger.LogInfo(fmt.Sprintf("Selected events successfully linked and added to  %s control event log", pyxis))
			printfln("selected events successfully linked and added to %s control event log", pyxis)

			return nil

		}

	}, cli.CommandArg{
		Name:     "pyxis",
		Required: true,
	}, cli.CommandArg{
		Name:     "modifier",
		Required: false,
	})

}
