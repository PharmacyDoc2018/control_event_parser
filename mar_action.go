package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/PharmacyDoc2018/pyxis_event_tracker/database"
)

type MarAction struct {
	SavedTime           time.Time //-- SAVED_TIME
	OrderNumber         string    //-- ORDER_MED_ID - convert to string
	MarAction           string    //-- FilteredMARAction
	DisplayName         string    //--DISPLAY_NAME
	MedicationID        string    //-- MEDICATION_ID convert to string
	UserID              string    //--SYSTEM_LOGIN
	UserName            string    //-- NAME
	DoseUnitDescription string    //-- CalcDoseUnitDescription
	Dose                float64   //-- CALC_MIN_DOSE
	MRN                 string    //-- PAT_MRN_ID
	PtName              string    //-- PAT_NAME
}

func (p *Process) parseMarActions(response []database.MarActionResponse) []MarAction {
	parsedMarActions := []MarAction{}

	for _, action := range response {
		marAction := MarAction{}

		if action.SavedTime.Valid {
			marAction.SavedTime = action.SavedTime.Time
		} else {
			marAction.SavedTime = time.Time{}
		}

		marAction.OrderNumber = strconv.Itoa(action.OrderMedId)

		if action.FilteredMarAction.Valid {
			marAction.MarAction = action.FilteredMarAction.String
		} else {
			marAction.MarAction = ""
		}

		if action.DisplayName.Valid {
			marAction.DisplayName = action.DisplayName.String
		} else {
			marAction.DisplayName = ""
		}

		if action.MedicationId.Valid {
			marAction.MedicationID = strconv.Itoa(int(action.MedicationId.Int64))
		} else {
			marAction.MedicationID = ""
		}

		if action.SystemLogin.Valid {
			marAction.UserID = action.SystemLogin.String
		} else {
			marAction.UserID = ""
		}

		if action.Name.Valid {
			marAction.UserName = action.Name.String
		} else {
			marAction.UserName = ""
		}

		if action.CalcDoseUnitDescription.Valid {
			splitDescription := strings.Split(action.CalcDoseUnitDescription.String, " ")
			marAction.DoseUnitDescription = splitDescription[len(splitDescription)-1]
		} else {
			marAction.DoseUnitDescription = ""
		}

		if action.CalcMinDose.Valid {
			if action.DispQty.Valid && action.DispQtyUnitName.Valid && action.Strength.Valid {
				marAction.Dose, _ = parseDose(action.DispQty.Float64, action.DispQtyUnitName.String, action.Strength.String)
			} else {
				marAction.Dose = action.CalcMinDose.Float64
			}
		} else {
			marAction.Dose = 0.0
		}

		if action.PatMRN.Valid {
			marAction.MRN = action.PatMRN.String
		} else {
			marAction.MRN = ""
		}

		if action.PatName.Valid {
			marAction.PtName = action.PatName.String
		} else {
			marAction.PtName = ""
		}

		parsedMarActions = append(parsedMarActions, marAction)
	}

	return parsedMarActions
}
