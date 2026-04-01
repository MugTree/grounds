package app

import (
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func pathValueAsIntOrErr(w http.ResponseWriter, r *http.Request, key string) (string, bool) {

	formVal := r.PathValue(key)

	if formVal == "" {
		errorHandler(w, r, fmt.Sprintf("http: incorrect path value %s on page %v", key, r.URL.Path))
		return "0", false
	}

	_, err := strconv.Atoi(formVal)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("http: incorrect path value %v, should be numeric - on page %v", formVal, r.URL.Path))
		return "0", false
	}

	return formVal, true

}

func formValueAsIntOrErr(w http.ResponseWriter, r *http.Request, key string) (string, bool) {

	formVal := r.FormValue(key)

	if formVal == "" {
		errorHandler(w, r, fmt.Sprintf("http: incorrect form value %s on page %v", key, r.URL.Path))
		return "0", false
	}

	_, err := strconv.Atoi(formVal)
	if err != nil {
		errorHandler(w, r, fmt.Sprintf("http: incorrect form value %v, should be numeric - on page %v", formVal, r.URL.Path))
		return "", false
	}

	return formVal, true

}

func parseMultipart(r *http.Request) (*http.Request, error) {

	ct := r.Header.Get("Content-Type")
	if strings.HasPrefix(ct, "multipart/form-data") {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			return r, err
		}
	} else {
		err := r.ParseForm()
		if err != nil {
			return r, err
		}
	}

	return r, nil
}

func setAriaValidity(val bool) string {
	if val {
		return "true"
	}
	return "false"
}

func hasDateError(input string) bool {
	_, err := time.Parse("2006-01-02", input)
	if err != nil {
		return true
	}
	return false
}

func hasTimeError(input string) bool {
	_, err := time.Parse("15:04", input)
	if err != nil {
		return true
	}
	return false
}

func hasNotesError(_ string) bool {
	return false
}

func validateVisit(r *http.Request) VisitVM {

	vm := VisitVM{}
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

	if hasDateError(r.FormValue("visit_date")) {
		vm.HasDateError = true
	}

	if hasTimeError(r.FormValue("visit_time")) {
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
