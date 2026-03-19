package app

import (
	"fmt"
	"os"
	"testing"
)

func Test_DecodeImage(t *testing.T) {

	f, err := os.Open("/Users/me/home/Dev/go-projects/blanchflowerguitars/raw_images/bolt_slim/1000034310.jpg")
	if err != nil {
		t.Error(err)
	}

	mimeType, ext, err := validateUpload(f)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(mimeType)
	fmt.Println(ext)

}
