package real_estate_api

type ZooplaClientInterface interface {
	SearchProperties(area string, minPrice, maxPrice, bedrooms int, propertyType string) ([]Property, error)
	TestApiConnection() error
}
