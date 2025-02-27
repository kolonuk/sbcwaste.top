// include all the required modules
package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// convertImageToBase64 fetches an image from a URL and converts it to a base64 data URI string
func convertImageToBase64URI(imageUrl string) (string, error) {
	// Fetch the image
	resp, err := http.Get(imageUrl)
	if err != nil {
		return "", fmt.Errorf("failed to fetch image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Read the image data
	imageData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read image data: %v", err)
	}

	// Determine the content type
	contentType := http.DetectContentType(imageData)

	// Encode the image data to base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)

	// Format as data URI
	base64URI := fmt.Sprintf("data:%s;base64,%s", contentType, base64Data)

	return base64URI, nil
}

func generateUID(eventTitle, eventStartDate, eventLocation string) string {
	// Concatenate the constant attributes
	concatenatedAttributes := eventTitle + eventStartDate + eventLocation

	// Hash the concatenated string
	hasher := sha256.New()
	hasher.Write([]byte(concatenatedAttributes))
	hash := hasher.Sum(nil)

	// Convert the hash to a hexadecimal string and format the UID
	uid := fmt.Sprintf("%x@sbcwaste.com", hash)

	return uid
}

func queryAddressAPI(url string, debuggingEnable bool) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if debuggingEnable {
		log.Printf("Response body: %s\n", string(body))
	}

	re := regexp.MustCompile(`\(([\s\S]*?)\);?$`)
	matches := re.FindSubmatch(body)
	if len(matches) < 2 {
		return "", errors.New("failed to extract JSON from JSONP response")
	}
	jsonString := string(matches[1])

	var addressResponse AddressResponse
	err = json.Unmarshal([]byte(jsonString), &addressResponse)
	if err != nil {
		return "", err
	}

	if len(addressResponse.Data) > 0 && len(addressResponse.Data[0]) > 2 {
		return addressResponse.Data[0][2], nil
	}

	return "", nil
}

func foldLine(s string) string {
	const maxLen = 74
	var result strings.Builder

	for x := 0; x < len(s); x += maxLen {
		// Determine the end of the current slice
		end := x + maxLen
		if end > len(s) {
			end = len(s)
		}
		// If this is not the first slice, a space at the start
		if x > 1 {
			result.WriteString(" ")
		}
		// Append the current slice to the result
		result.WriteString(s[x:end])
		// If this is not the last slice, add a CRLF
		if end != len(s) {
			result.WriteString("\r\n")
		}
	}

	return result.String()
}
