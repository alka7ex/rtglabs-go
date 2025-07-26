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
			typesense.WithServer("http://"+host+":80"), // <-- Issue is likely here
			typesense.WithAPIKey(apiKey),
		),
	}
}

