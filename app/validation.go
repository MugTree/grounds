package app

import (
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"time"
)

func isValidDate(input string) bool {
	_, err := time.Parse("2006-01-02", input)
	if err != nil {
		return false
	}
	return true
}

func isValidTime(input string) bool {
	_, err := time.Parse("15:04", input)
	if err != nil {
		return false
	}
	return true
}

func areValidNotes(_ string) bool {
	return true
}

func validateVisitSubmission(r *http.Request) visitVM {

	vm := visitVM{}
	/* in a realistic scenario we would validate
	everything
	not just date and time
	*/

	vm.CustomerId = r.FormValue("customer_id")
	vm.CustomerName = r.FormValue("customer_name")
	vm.LocationId = r.FormValue("location_id")
	vm.LocationName = r.FormValue("location_name")
	vm.Date = r.FormValue("visit_date")
	vm.Time = r.FormValue("visit_date")
	vm.Duration = r.FormValue("visit_duration")
	vm.IsSubmission = true

	fmt.Println("from form - visit date", "'"+vm.Date+"'")
	fmt.Println("from form - visit time", "'"+vm.Time+"'")

	if !isValidDate(r.FormValue("visit_date")) {
		vm.HasDateError = true
	}

	if !isValidTime(r.FormValue("visit_time")) {
		vm.HasTimeError = true
	}

	vm.HasNotesError = false

	return vm

}

// image.DecodeConfig proves that the file bytes are decodable as an image format the go actually understands
// only allows jpges atm
func validateUpload(file io.ReadSeeker) (string, string, error) {
	buf := make([]byte, 512)

	n, err := file.Read(buf)
	if err != nil {
		return "", "", err
	}

	mimeType := http.DetectContentType(buf[:n])

	_, err = file.Seek(0, 0)
	if err != nil {
		return "", "", err
	}

	_, _, err = image.DecodeConfig(file)
	if err != nil {
		return "", "", errors.New("invalid image data")
	}

	_, err = file.Seek(0, 0)
	if err != nil {
		return "", "", err
	}

	if mimeType != "image/jpeg" {
		return "", "", errors.New("unsupported image type")
	}

	return mimeType, ".jpg", nil
}
