package schemas

import (
	"context"
	"log"
	"strings"

	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate/entities/models"
)

// Try to create the passed schema, ignore if
// the schema already exists
func InitSchema(client *weaviate.Client, class *models.Class) error {
	err := client.Schema().ClassCreator().WithClass(class).Do(context.Background())

	if err != nil {
		// 422 error returned for a schema that already exists
		// so if anything else is returned there may be an
		// actual issue
		if strings.Contains(err.Error(), "422") {
			log.Println("Schema already exists for class", class.Class)
		} else {
			return err
		}
	} else {
		log.Println("Creating schema for class", class.Class)
	}

	return nil
}
