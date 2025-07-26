package provider

import (
	"github.com/typesense/typesense-go/v3/typesense"
)

type TypesenseClient struct {
	Client *typesense.Client
}

func NewTypesenseClient(host, apiKey string) *TypesenseClient {
	return &TypesenseClient{
		Client: typesense.NewClient(
			typesense.WithServer("https://"+host+":443"),
			typesense.WithAPIKey(apiKey),
		),
	}
}
