package app

type dateSignals struct {
	VisitDate string `json:"visit_date"`
}

type timeSignals struct {
	VisitTime string `json:"visit_time"`
}

type notesSignals struct {
	VisitNotes string `json:"visit_notes"`
}

type VisitVM struct {
	Date         string
	Time         string
	Duration     string
	Notes        string
	CustomerId   int64
	CustomerName string
	LocationName string
	LocationId   int64
	IsComplete   bool
	IsSubmission bool
	VisitVMErrors
}

func (v VisitVM) HasErrors() bool {
	if v.HasDateError || v.HasTimeError {
		return true
	}
	return false
}

type VisitVMErrors struct {
	HasTimeError  bool
	HasDateError  bool
	HasNotesError bool
}
