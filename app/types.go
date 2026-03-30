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

// type HomepageVm struct {
// 	SelectedCustomer int
// 	SelectedLocation int
// 	ShowLocations    bool
// 	Customers        []Customer
// 	Locations        []Location
// 	IsValid          bool
// }

// type Journey struct {
// 	CustomerID string `json:"customerId,omitempty"`
// 	LocationID string `json:"locationId,omitempty"`
// }

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
