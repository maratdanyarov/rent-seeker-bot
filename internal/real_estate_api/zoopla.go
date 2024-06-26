package real_estate_api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	tokenURL     = "https://api.alto.zoopladev.co.uk/token"
	inventoryURL = "https://api.alto.zoopladev.co.uk/inventory"
)

type ZooplaClient struct {
	ClientID     string
	ClientSecret string
	AgencyRef    string
	Token        string
	TokenExpiry  time.Time
}

type Property struct {
	ID          string `json:"listing_id"`
	Address     string `json:"address"`
	Price       int    `json:"price"`
	Bedrooms    int    `json:"num_bedrooms"`
	Description string `json:"description"`
	URL         string `json:"details_url"`
	// Add more fields as needed
}

func NewZooplaClient(clientID, clientSecret, agencyRef string) *ZooplaClient {
	return &ZooplaClient{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		AgencyRef:    agencyRef,
	}
}

func (c *ZooplaClient) getToken() error {
	if c.Token != "" && time.Now().Before(c.TokenExpiry) {
		return nil
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(c.ClientID, c.ClientSecret)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	c.Token = result.AccessToken
	c.TokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)

	return nil
}

func (c *ZooplaClient) SearchProperties(area string, minPrice, maxPrice, bedrooms int, propertyType string) ([]Property, error) {
	log.Printf("Searching for properties: area=%s, minPrice=%d, maxPrice=%d, bedrooms=%d, type=%s",
		area, minPrice, maxPrice, bedrooms, propertyType)

	if err := c.getToken(); err != nil {
		return nil, fmt.Errorf("error getting token %v", err)
	}

	// Construct the API URL with the search parameters
	query := url.Values{}
	query.Add("address", area)
	query.Add("minimum_price", fmt.Sprintf("%d", minPrice))
	query.Add("maximum_price", fmt.Sprintf("%d", maxPrice))
	query.Add("minimum_beds", fmt.Sprintf("%d", bedrooms))
	query.Add("property_type", propertyType)

	req, err := http.NewRequest("GET", inventoryURL+"?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("AgencyRef", c.AgencyRef)
	req.Header.Set("Authorization", "Bearer "+c.Token)

	log.Printf("Sending request to Zoopla API: %s", req.URL.String())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	log.Printf("Received response from Zoopla API. Status: %s", resp.Status)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Properties []Property `json:"properties"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error unmarshaling JSON: %v. Body: %s", err, string(body))
		return nil, err
	}

	log.Printf("Successfully parsed %d properties from Zoopla API response", len(result.Properties))

	return result.Properties, nil
}
func (c *ZooplaClient) TestApiConnection() error {
	// Test if we can get a token
	err := c.getToken()
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}
	// Try a simple search
	properties, err := c.SearchProperties("London", 1000, 2000, 2, "Flat")
	if err != nil {
		return fmt.Errorf("failed to search propeties: %w", err)
	}

	// print the first of the results
	if len(properties) > 0 {
		log.Printf("Successfully retrieved %d properties. First property: %+v", len(properties), properties[0])
	} else {
		log.Println("Search successful, but no properties found.")
	}

	return nil
}
