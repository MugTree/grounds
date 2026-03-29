package app

type Customer struct {
	Id   int    `db:"id"`
	Name string `db:"name"`
}

type VisitCompleteVm struct {
	LocationName string
	CustomerName string
	EmployeeName string
	VisitId      string
	Time         string
	Date         string
	Duration     string
	ImagePaths   []string
}

type dateSignals struct {
	VisitDate string `json:"visit_date"`
}

type getLocSignals struct {
	CustomerId string `json:"customerId"`
}

type homePageSignals struct {
	CustomerId int `json:"customerId"`
	LocationId int `json:"locationId"`
}

type HomepageVm struct {
	SelectedCustomer int
	SelectedLocation int
	ShowLocations    bool
	Customers        []Customer
	Locations        []Location
	IsValid          bool
}

type notesSignals struct {
	VisitNotes string `json:"visit_notes"`
}

type Journey struct {
	CustomerID string `json:"customerId,omitempty"`
	LocationID string `json:"locationId,omitempty"`
}

type Location struct {
	Id         string `db:"id"`
	Name       string `db:"name"`
	CustomerId string `db:"customer_id"`
}

type PickCustomerVm struct {
	Customers []Customer
	HasError  bool
	//PreviousVisits []
}

type PickLocationVm struct {
	CustomerId   string
	CustomerName string
	Locations    []Location
	HasError     bool
}

type timeSignals struct {
	VisitTime string `json:"visit_time"`
}

type VisitVM struct {
	Date         string
	Time         string
	Duration     string
	Notes        string
	CustomerId   string
	CustomerName string
	LocationName string
	LocationId   string
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

type Visit struct {
	Id         int `db:"id"`
	EmployeeId int `db:"employee_id"`
	LocationId int `db:"location_id"`
}
