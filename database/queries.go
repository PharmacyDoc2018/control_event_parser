package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type MedicationName struct {
	MedDisplayName string
}

const listControlTwoMedsByDevice = `
SELECT DISTINCT i.MedDisplayName
FROM PYX.fctItemTransaction t
INNER JOIN PYX.dimDispensingDevice d
ON t.UserAtDispensingDeviceKey = d.DispensingDeviceKey
INNER JOIN PYX.dimItem i
ON t.ItemKey = i.ItemKey
WHERE d.DispensingDeviceName = @device
AND i.MedClassCode = '2'
`

func (q *Queries) ListControlTwoMedsByDevice(ctx context.Context, device string) ([]MedicationName, error) {
	rows, err := q.db.QueryContext(ctx, listControlTwoMedsByDevice, sql.Named("device", device))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []MedicationName
	for rows.Next() {
		var i MedicationName
		if err := rows.Scan(
			&i.MedDisplayName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getPyxisEventsForDeviceByDateRange = `
SELECT
	t.ItemTransactionKey,
	u.FullName AS UserName,
	u.UserID,
	space.StorageSpaceShortName AS StorageSpace,
	i.ItemID,
	i.MedClassCode,
	i.MedDisplayName,
	tc.TransactionType,
	t.TransactionLocalDateTime,
	t.EnteredQuantity,
	t.EnteredUOMDisplayCode,
	CAST(
		CASE
			WHEN t.EnteredUOMDisplayCode IS NULL
				THEN NULL
			WHEN t.EnteredUOMDisplayCode = 'Dosage Form'
				THEN t.EnteredQuantity * i.StrengthAmount
			WHEN t.EnteredUOMDisplayCode = i.VolumeUnit_DisplayCode
				THEN (t.EnteredQuantity * i.StrengthAmount) / i.TotalVolumeAmount
			ELSE
				t.EnteredQuantity
		END
		AS numeric(18,4)
	) AS AmountReferenced,
	i.StrengthUnit_DisplayCode AS AmountReferencedUnits,
	t.BegInventory,
	t.EndInventory,
	w.FullName AS WitnessName,
	w.UserID AS WitnessID,
	p.PatientID AS MRN
FROM PYX.fctItemTransaction t
INNER JOIN PYX.dimDispensingDevice d
	ON t.UserAtDispensingDeviceKey = d.DispensingDeviceKey
INNER JOIN PYX.dimItem i
	ON t.ItemKey = i.ItemKey
INNER JOIN dbo.dimTime time
	ON t.Tx_TimeKey = time.DimTime_key
LEFT JOIN PYX.fctItemEncounters e
	ON t.EncounterKey = e.EncounterKey
INNER JOIN PYX.dimUserAccount u
	ON t.UserAccountKey = u.UserAccountKey
INNER JOIN PYX.dimUserAccount w
	ON t.WitnessAccountKey = w.UserAccountKey
INNER JOIN PYX.dimTransactionCharacteristics tc
	ON t.sk_dim_TxChar = tc.sk_dim_TxChar
LEFT JOIN PYX.fctPatient p
	ON e.PatientKey = p.PatientKey
LEFT JOIN PYX.dimStorageSpace space
	ON t.StorageSpaceKey = space.StorageSpaceKey
WHERE d.DispensingDeviceName = @device
AND t.TransactionLocalDateTime > @start
AND t.TransactionLocalDateTime < @end
AND NOT (
	tc.TransactionType = 'Remove'
	AND
	t.BegInventory IS NULL
)
ORDER BY
	t.Tx_Date,
	time.string_representation_24,
	e.LastItemTransactionLocalDateTime;
`

type PyxisEventResponse struct {
	ItemTransactionKey    uuid.UUID
	UserName              sql.NullString
	UserID                sql.NullString
	StorageSpace          sql.NullString
	ItemID                sql.NullString
	MedClassCode          sql.NullString
	MedDisplayName        sql.NullString
	TransactionType       sql.NullString
	TxDateTime            sql.NullTime
	EnteredQuantity       sql.NullFloat64
	EnteredUOMDisplayCode sql.NullString
	AmountReferenced      sql.NullFloat64
	AmountReferencedUnits sql.NullString
	BegInventory          sql.NullFloat64
	EndInventory          sql.NullFloat64
	WitnessName           sql.NullString
	WitnessID             sql.NullString
	MRN                   sql.NullString
}

type GetPyxisEventsForDeviceByDateRangeParams struct {
	Device string
	Start  time.Time
	End    time.Time
}

func (q *Queries) GetPyxisEventsForDeviceByDateRange(ctx context.Context, arg GetPyxisEventsForDeviceByDateRangeParams) ([]PyxisEventResponse, error) {
	rows, err := q.db.QueryContext(ctx, getPyxisEventsForDeviceByDateRange, sql.Named("device", arg.Device), sql.Named("start", arg.Start), sql.Named("end", arg.End))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PyxisEventResponse
	for rows.Next() {
		var i PyxisEventResponse
		if err := rows.Scan(
			&i.ItemTransactionKey,
			&i.UserName,
			&i.UserID,
			&i.StorageSpace,
			&i.ItemID,
			&i.MedClassCode,
			&i.MedDisplayName,
			&i.TransactionType,
			&i.TxDateTime,
			&i.EnteredQuantity,
			&i.EnteredUOMDisplayCode,
			&i.AmountReferenced,
			&i.AmountReferencedUnits,
			&i.BegInventory,
			&i.EndInventory,
			&i.WitnessName,
			&i.WitnessID,
			&i.MRN,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil

}

const getPyxisEventByItemTransactionKey = `
SELECT
	t.ItemTransactionKey,
	u.FullName AS UserName,
	u.UserID,
	space.StorageSpaceShortName AS StorageSpace,
	i.ItemID,
	i.MedClassCode,
	i.MedDisplayName,
	tc.TransactionType,
	t.TransactionLocalDateTime,
	t.EnteredQuantity,
	t.EnteredUOMDisplayCode,
	CAST(
		CASE
			WHEN t.EnteredUOMDisplayCode IS NULL
				THEN NULL
			WHEN t.EnteredUOMDisplayCode = 'Dosage Form'
				THEN t.EnteredQuantity * i.StrengthAmount
			WHEN t.EnteredUOMDisplayCode = i.VolumeUnit_DisplayCode
				THEN (t.EnteredQuantity * i.StrengthAmount) / i.TotalVolumeAmount
			ELSE
				t.EnteredQuantity
		END
		AS numeric(18,4)
	) AS AmountReferenced,
	i.StrengthUnit_DisplayCode AS AmountReferencedUnits,
	t.BegInventory,
	t.EndInventory,
	w.FullName AS WitnessName,
	w.UserID AS WitnessID,
	p.PatientID AS MRN
FROM PYX.fctItemTransaction t
INNER JOIN PYX.dimDispensingDevice d
	ON t.UserAtDispensingDeviceKey = d.DispensingDeviceKey
INNER JOIN PYX.dimItem i
	ON t.ItemKey = i.ItemKey
INNER JOIN dbo.dimTime time
	ON t.Tx_TimeKey = time.DimTime_key
LEFT JOIN PYX.fctItemEncounters e
	ON t.EncounterKey = e.EncounterKey
INNER JOIN PYX.dimUserAccount u
	ON t.UserAccountKey = u.UserAccountKey
INNER JOIN PYX.dimUserAccount w
	ON t.WitnessAccountKey = w.UserAccountKey
INNER JOIN PYX.dimTransactionCharacteristics tc
	ON t.sk_dim_TxChar = tc.sk_dim_TxChar
LEFT JOIN PYX.fctPatient p
	ON e.PatientKey = p.PatientKey
LEFT JOIN PYX.dimStorageSpace space
	ON t.StorageSpaceKey = space.StorageSpaceKey
WHERE t.ItemTransactionKey = @id;
`

func (q *Queries) GetPyxisEventByItemTransactionKey(ctx context.Context, id string) (PyxisEventResponse, error) {
	row, err := q.db.QueryContext(ctx, getPyxisEventByItemTransactionKey, sql.Named("id", id))
	if err != nil {
		return PyxisEventResponse{}, err
	}
	defer row.Close()

	var item PyxisEventResponse
	if !row.Next() {
		return PyxisEventResponse{}, fmt.Errorf("no result found")
	}
	if err := row.Scan(
		&item.ItemTransactionKey,
		&item.UserName,
		&item.UserID,
		&item.StorageSpace,
		&item.ItemID,
		&item.MedClassCode,
		&item.MedDisplayName,
		&item.TransactionType,
		&item.TxDateTime,
		&item.EnteredQuantity,
		&item.EnteredUOMDisplayCode,
		&item.AmountReferenced,
		&item.AmountReferencedUnits,
		&item.BegInventory,
		&item.EndInventory,
		&item.WitnessName,
		&item.WitnessID,
		&item.MRN,
	); err != nil {
		return PyxisEventResponse{}, err
	}

	if err := row.Close(); err != nil {
		return PyxisEventResponse{}, err
	}
	if err := row.Err(); err != nil {
		return PyxisEventResponse{}, err
	}
	return item, nil

}

type MarActionResponse struct {
	SavedTime               sql.NullTime
	OrderMedId              int
	FilteredMarAction       sql.NullString
	DisplayName             sql.NullString
	MedicationId            sql.NullInt64
	SystemLogin             sql.NullString
	Name                    sql.NullString
	CalcDoseUnitDescription sql.NullString
	CalcMinDose             sql.NullFloat64
	DispQty                 sql.NullFloat64
	DispQtyUnitName         sql.NullString
	Strength                sql.NullString
	PatMRN                  sql.NullString
	PatName                 sql.NullString
}

type GetMarAdminActionsByPatientDayMedIDsParams struct {
	Date    time.Time
	DeptIDs []string
	Mrn     string
	MedIDs  []string
}

func (q *Queries) GetMarAdminActionsByPatientDayMedIDs(ctx context.Context, arg GetMarAdminActionsByPatientDayMedIDsParams) ([]MarActionResponse, error) {
	deptIdPlaceholders := make([]string, len(arg.DeptIDs))
	deptIdArgs := make([]any, len(arg.DeptIDs))

	medIdPlaceholders := make([]string, len(arg.MedIDs))
	medIDArgs := make([]any, len(arg.MedIDs))

	h, m, s := arg.Date.Clock()
	if h != 0 || m != 0 || s != 0 {
		return nil, fmt.Errorf("error. date must be set to midnight of the selected day")
	}

	endDate := arg.Date.Add(24 * time.Hour)

	allArgs := []any{
		sql.Named("start", arg.Date),
		sql.Named("end", endDate),
		sql.Named("mrn", arg.Mrn),
	}

	for i, deptID := range arg.DeptIDs {
		deptIdPlaceholders[i] = fmt.Sprintf("@d%d", i)
		deptIdArgs[i] = sql.Named(fmt.Sprintf("d%d", i), deptID)
	}
	allArgs = append(allArgs, deptIdArgs...)

	for i, medID := range arg.MedIDs {
		medIdPlaceholders[i] = fmt.Sprintf("@p%d", i)
		medIDArgs[i] = sql.Named(fmt.Sprintf("p%d", i), medID)
	}
	allArgs = append(allArgs, medIDArgs...)

	query := fmt.Sprintf(`
SELECT
	ma.SAVED_TIME,
	ma.ORDER_MED_ID,
	mc.FilteredMARAction,
	o.DISPLAY_NAME, 
	o.MEDICATION_ID,
	u.SYSTEM_LOGIN,
	u.NAME,
	o.CalcDoseUnitDescription,
	o.CALC_MIN_DOSE,
	mam.DISP_QTY,
	mu.NAME,
	med.STRENGTH,
	pat.PAT_MRN_ID,
	pat.PAT_NAME
FROM dbo.fctMARActions ma
INNER JOIN dbo.dimMARCharacteristics mc
	ON ma.sk_dim_MARChar = mc.sk_dim_MARChar
INNER JOIN dbo.fctMARActionMedications mam
	ON ma.ORDER_MED_ID = mam.ORDER_MED_ID
	AND ma.LINE = mam.MAR_LINE
INNER JOIN dbo.dimMedUnit mu
	ON mam.DISP_QTYUNIT_C = mu.DISP_QTYUNIT_C
INNER JOIN dbo.fctORDER_MED o
	ON ma.ORDER_MED_ID = o.ORDER_MED_ID
INNER JOIN dbo.dimMedication med
	ON o.MEDICATION_ID = med.MEDICATION_ID
INNER JOIN dbo.dimUsers u
	ON ma.MAR_ActionUser = u.User_ID
INNER JOIN dbo.dimPatient pat
	ON pat.PAT_ID = o.PAT_ID
WHERE mc.FilteredMARAction IN (
	'Given',
	'New Bag'
)
AND ma.SAVED_TIME >= @start
AND ma.SAVED_TIME < @end
AND ma.MAR_ADMIN_DEPT_ID IN (%s)
AND pat.PAT_MRN_ID = @mrn
AND o.MEDICATION_ID IN (%s)
ORDER BY ma.SAVED_TIME;
`, strings.Join(deptIdPlaceholders, ","),
		strings.Join(medIdPlaceholders, ","))

	rows, err := q.db.QueryContext(ctx, query, allArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MarActionResponse
	for rows.Next() {
		var i MarActionResponse
		if err := rows.Scan(
			&i.SavedTime,
			&i.OrderMedId,
			&i.FilteredMarAction,
			&i.DisplayName,
			&i.MedicationId,
			&i.SystemLogin,
			&i.Name,
			&i.CalcDoseUnitDescription,
			&i.CalcMinDose,
			&i.DispQty,
			&i.DispQtyUnitName,
			&i.Strength,
			&i.PatMRN,
			&i.PatName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil

}

type GetMarAdminActionsByPatientsDaysMedIDsParams struct {
	DateStart time.Time
	DateEnd   time.Time
	DeptIDs   []string
	Mrns      []string
	MedIDs    []string
}

func (q *Queries) GetMarAdminActionsByPatientsDaysMedIDs(ctx context.Context, arg GetMarAdminActionsByPatientsDaysMedIDsParams) ([]MarActionResponse, error) {
	deptIdPlaceholders := make([]string, len(arg.DeptIDs))
	deptIdArgs := make([]any, len(arg.DeptIDs))

	mrnPlaceholders := make([]string, len(arg.Mrns))
	mrnArgs := make([]any, len(arg.Mrns))

	medIdPlaceholders := make([]string, len(arg.MedIDs))
	medIDArgs := make([]any, len(arg.MedIDs))

	h, m, s := arg.DateStart.Clock()
	if h != 0 || m != 0 || s != 0 {
		return nil, fmt.Errorf("error. date must be set to midnight of the first selected day")
	}

	h, m, s = arg.DateEnd.Clock()
	if h != 23 || m != 59 || s != 59 {
		return nil, fmt.Errorf("error. date must be set to 23:59:59 of the last selected day")
	}

	allArgs := []any{
		sql.Named("start", arg.DateStart),
		sql.Named("end", arg.DateEnd),
	}

	for i, deptID := range arg.DeptIDs {
		deptIdPlaceholders[i] = fmt.Sprintf("@d%d", i)
		deptIdArgs[i] = sql.Named(fmt.Sprintf("d%d", i), deptID)
	}
	allArgs = append(allArgs, deptIdArgs...)

	for i, mrn := range arg.Mrns {
		mrnPlaceholders[i] = fmt.Sprintf("@m%d", i)
		mrnArgs[i] = sql.Named(fmt.Sprintf("m%d", i), mrn)
	}
	allArgs = append(allArgs, mrnArgs...)

	for i, medID := range arg.MedIDs {
		medIdPlaceholders[i] = fmt.Sprintf("@p%d", i)
		medIDArgs[i] = sql.Named(fmt.Sprintf("p%d", i), medID)
	}
	allArgs = append(allArgs, medIDArgs...)

	query := fmt.Sprintf(`
SELECT
	ma.SAVED_TIME,
	ma.ORDER_MED_ID,
	mc.FilteredMARAction,
	o.DISPLAY_NAME, 
	o.MEDICATION_ID,
	u.SYSTEM_LOGIN,
	u.NAME,
	o.CalcDoseUnitDescription,
	o.CALC_MIN_DOSE,
	mam.DISP_QTY,
	mu.NAME,
	med.STRENGTH,
	pat.PAT_MRN_ID,
	pat.PAT_NAME
FROM dbo.fctMARActions ma
INNER JOIN dbo.dimMARCharacteristics mc
	ON ma.sk_dim_MARChar = mc.sk_dim_MARChar
INNER JOIN dbo.fctMARActionMedications mam
	ON ma.ORDER_MED_ID = mam.ORDER_MED_ID
	AND ma.LINE = mam.MAR_LINE
INNER JOIN dbo.dimMedUnit mu
	ON mam.DISP_QTYUNIT_C = mu.DISP_QTYUNIT_C
INNER JOIN dbo.fctORDER_MED o
	ON ma.ORDER_MED_ID = o.ORDER_MED_ID
INNER JOIN dbo.dimMedication med
	ON o.MEDICATION_ID = med.MEDICATION_ID
INNER JOIN dbo.dimUsers u
	ON ma.MAR_ActionUser = u.User_ID
INNER JOIN dbo.dimPatient pat
	ON pat.PAT_ID = o.PAT_ID
WHERE mc.FilteredMARAction IN (
	'Given',
	'New Bag'
)
AND ma.SAVED_TIME >= @start
AND ma.SAVED_TIME < @end
AND ma.MAR_ADMIN_DEPT_ID IN (%s)
AND pat.PAT_MRN_ID IN (%s) 
AND o.MEDICATION_ID IN (%s)
ORDER BY ma.SAVED_TIME;
`, strings.Join(deptIdPlaceholders, ","),
		strings.Join(mrnPlaceholders, ","),
		strings.Join(medIdPlaceholders, ","))

	rows, err := q.db.QueryContext(ctx, query, allArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MarActionResponse
	for rows.Next() {
		var i MarActionResponse
		if err := rows.Scan(
			&i.SavedTime,
			&i.OrderMedId,
			&i.FilteredMarAction,
			&i.DisplayName,
			&i.MedicationId,
			&i.SystemLogin,
			&i.Name,
			&i.CalcDoseUnitDescription,
			&i.CalcMinDose,
			&i.DispQty,
			&i.DispQtyUnitName,
			&i.Strength,
			&i.PatMRN,
			&i.PatName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil

}

const getMarActionByOrderNumber = `
SELECT
	ma.SAVED_TIME,
	ma.ORDER_MED_ID,
	mc.FilteredMARAction,
	o.DISPLAY_NAME, 
	o.MEDICATION_ID,
	u.SYSTEM_LOGIN,
	u.NAME,
	o.CalcDoseUnitDescription,
	o.CALC_MIN_DOSE,
	mam.DISP_QTY,
	mu.NAME,
	med.STRENGTH,
	pat.PAT_MRN_ID,
	pat.PAT_NAME
FROM dbo.fctMARActions ma
INNER JOIN dbo.dimMARCharacteristics mc
	ON ma.sk_dim_MARChar = mc.sk_dim_MARChar
INNER JOIN dbo.fctMARActionMedications mam
	ON ma.ORDER_MED_ID = mam.ORDER_MED_ID
	AND ma.LINE = mam.MAR_LINE
INNER JOIN dbo.dimMedUnit mu
	ON mam.DISP_QTYUNIT_C = mu.DISP_QTYUNIT_C
INNER JOIN dbo.fctORDER_MED o
	ON ma.ORDER_MED_ID = o.ORDER_MED_ID
INNER JOIN dbo.dimMedication med
	ON o.MEDICATION_ID = med.MEDICATION_ID
INNER JOIN dbo.dimUsers u
	ON ma.MAR_ActionUser = u.User_ID
INNER JOIN dbo.dimPatient pat
	ON pat.PAT_ID = o.PAT_ID
WHERE mc.FilteredMARAction IN (
	'Given',
	'New Bag'
)
AND ma.ORDER_MED_ID = @order
ORDER BY ma.SAVED_TIME;
`

func (q *Queries) GetMarActionsByOrderNumber(ctx context.Context, orderNumber string) ([]MarActionResponse, error) {
	rows, err := q.db.QueryContext(ctx, getMarActionByOrderNumber, sql.Named("order", orderNumber))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []MarActionResponse
	for rows.Next() {
		var i MarActionResponse
		if err := rows.Scan(
			&i.SavedTime,
			&i.OrderMedId,
			&i.FilteredMarAction,
			&i.DisplayName,
			&i.MedicationId,
			&i.SystemLogin,
			&i.Name,
			&i.CalcDoseUnitDescription,
			&i.CalcMinDose,
			&i.DispQty,
			&i.DispQtyUnitName,
			&i.Strength,
			&i.PatMRN,
			&i.PatName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}

	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
