package real_estate_api

import (
	"fmt"
	"math/rand"
	"strings"
)

type MockZooplaClient struct {
}

func NewMockZooplaClient() *MockZooplaClient {
	return &MockZooplaClient{}
}

func (c *MockZooplaClient) SearchProperties(area string, minPrice, maxPrice, bedrooms int, propertyType string) ([]Property, error) {
	var properties []Property
	numProperties := rand.Intn(5) + 1 // Return 1-5 properties

	for i := 0; i < numProperties; i++ {
		price := rand.Intn(maxPrice-minPrice+1) + minPrice
		property := Property{
			ID:       fmt.Sprintf("%d", rand.Intn(10000)),
			Address:  fmt.Sprintf("%d %s, %s", rand.Intn(100)+1, randomStreet(), area),
			Price:    price,
			Bedrooms: bedrooms,
			Description: fmt.Sprintf("A lovely %d bedroom %s in %s. This property is %s and available for £%d per month.",
				bedrooms, strings.ToLower(propertyType), area, randomCondition(), price),
		}
		properties = append(properties, property)
	}

	return properties, nil
}

func (c *MockZooplaClient) TestApiConnection() error {
	// Always return success for the mock
	return nil
}

func randomStreet() string {
	streets := []string{"High Street", "Church Road", "Main Street", "Park Road", "London Road"}
	return streets[rand.Intn(len(streets))]
}

func randomCondition() string {
	conditions := []string{"well-maintained", "newly renovated", "in good condition", "charming"}
	return conditions[rand.Intn(len(conditions))]
}
