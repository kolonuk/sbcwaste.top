// include all the required modules
package main

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// newSafeClient creates an http.Client that prevents requests to private networks.
func newSafeClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, _, err := net.SplitHostPort(addr)
				if err != nil {
					// This should not happen for http/https requests, as transport adds the port.
					// If it does, we can't safely check the host.
					return nil, fmt.Errorf("cannot split host/port: %w", err)
				}

				ips, err := net.LookupIP(host)
				if err != nil {
					return nil, fmt.Errorf("dns lookup failed for %s: %w", host, err)
				}

				for _, ip := range ips {
					if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
						return nil, fmt.Errorf("refused to connect to private/local address: %s (%s)", ip, host)
					}
				}

				// The address is safe, proceed with the default dialer
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
		Timeout: 60 * time.Second,
	}
}

var HTTPClient = newSafeClient()
var InsecureHTTPClient = newInsecureSafeClient()

// newInsecureSafeClient creates an http.Client that allows insecure connections.
func newInsecureSafeClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // #nosec G402
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				host, _, err := net.SplitHostPort(addr)
				if err != nil {
					// This should not happen for http/https requests, as transport adds the port.
					// If it does, we can't safely check the host.
					return nil, fmt.Errorf("cannot split host/port: %w", err)
				}

				ips, err := net.LookupIP(host)
				if err != nil {
					return nil, fmt.Errorf("dns lookup failed for %s: %w", host, err)
				}

				for _, ip := range ips {
					if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
						return nil, fmt.Errorf("refused to connect to private/local address: %s (%s)", ip, host)
					}
				}

				// The address is safe, proceed with the default dialer
				var d net.Dialer
				return d.DialContext(ctx, network, addr)
			},
		},
		Timeout: 60 * time.Second,
	}
}

// convertImageToBase64 fetches an image from a URL and converts it to a base64 data URI string
var convertImageToBase64URI = func(imageUrl string) (string, error) {
	// Fetch the image
	// #nosec G107 - Mitigated by using a safe http client.
	resp, err := HTTPClient.Get(imageUrl)
	if err != nil {
		return "", fmt.Errorf("failed to fetch image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	// Read the image data
	imageData, err := io.ReadAll(resp.Body)
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

// AddressResponse defines the structure for the address API response.
type AddressResponse struct {
	Data [][]string `json:"data"`
}

var fetchAddressData = func(client *http.Client, url string) (*AddressResponse, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get address data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read address data response body: %w", err)
	}

	// The response is JSONP, so we need to strip the callback function
	jsonp := string(body)
	// Find the first '(' and the last ')'
	start := strings.Index(jsonp, "(")
	end := strings.LastIndex(jsonp, ")")

	if start == -1 || end == -1 || start >= end {
		return nil, fmt.Errorf("invalid JSONP format: %s", jsonp)
	}

	jsonp = jsonp[start+1 : end]

	var addressResponse AddressResponse
	if err := json.Unmarshal([]byte(jsonp), &addressResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal address data: %w", err)
	}

	return &addressResponse, nil
}

func getAddressFromUPRN(client *http.Client, uprn string, debuggingEnable bool) (string, error) {
	url := "https://maps.swindon.gov.uk/getdata.aspx?callback=jQuery16406504322666596749_1721033956585&type=jsonp&service=LocationSearch&RequestType=LocationSearch&location=" + uprn + "&pagesize=13&startnum=1&gettotals=false&axuid=1721033978935&mapsource=mapsources/MyHouse&_=1721033978935"
	if debuggingEnable {
		log.Printf("Querying address API with URL: %s", strings.ReplaceAll(url, "\n", "")) // #nosec G706 -- url built from regex-validated UPRN and hardcoded base, newlines stripped
	}

	addressResponse, err := fetchAddressData(client, url)
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
