package db

import (
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
)

func InitCollection(app *pocketbase.PocketBase) error {
	// Check if collection exists
	// If not, create collection

	// Check if "anime_info" collection exists
	animeInfoCollection, _ := app.Dao().FindCollectionByNameOrId("anime_info")
	if animeInfoCollection == nil {
		if err := createAnimeInfoCollection(app); err != nil {
			return err
		}
	}

	// Check if "anime_subtitle" collection exists
	animeSubtitleCollection, _ := app.Dao().FindCollectionByNameOrId("anime_subtitle")
	if animeSubtitleCollection == nil {
		if err := createAnimeSubtitleCollection(app); err != nil {
			return err
		}
	}

	return nil
}

func createAnimeInfoCollection(app *pocketbase.PocketBase) error {
	collection := &models.Collection{
		Name:       "anime_info",
		Type:       models.CollectionTypeBase,
		ListRule:   nil,
		ViewRule:   nil,
		CreateRule: nil,
		UpdateRule: nil,
		DeleteRule: nil,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:     "week",
				Type:     schema.FieldTypeText,
				Required: true,
			},
			&schema.SchemaField{
				Name:     "anime_no",
				Type:     schema.FieldTypeNumber,
				Required: true,
			},
			&schema.SchemaField{
				Name: "status",
				Type: schema.FieldTypeText,
			},
			&schema.SchemaField{
				Name: "time",
				Type: schema.FieldTypeText,
			},
			&schema.SchemaField{
				Name:     "subject",
				Type:     schema.FieldTypeText,
				Required: true,
			},
			&schema.SchemaField{
				Name: "genres",
				Type: schema.FieldTypeText,
			},
			&schema.SchemaField{
				Name: "caption_count",
				Type: schema.FieldTypeNumber,
			},
			&schema.SchemaField{
				Name: "start_date",
				Type: schema.FieldTypeText,
			},
			&schema.SchemaField{
				Name: "end_date",
				Type: schema.FieldTypeText,
			},
			&schema.SchemaField{
				Name: "website",
				Type: schema.FieldTypeText,
			},
			&schema.SchemaField{
				Name: "recent_episode",
				Type: schema.FieldTypeNumber,
			},
		),
	}

	if err := app.Dao().SaveCollection(collection); err != nil {
		return err
	}

	return nil
}

func createAnimeSubtitleCollection(app *pocketbase.PocketBase) error {
	collection := &models.Collection{
		Name:       "anime_subtitle",
		Type:       models.CollectionTypeBase,
		ListRule:   nil,
		ViewRule:   nil,
		CreateRule: nil,
		UpdateRule: nil,
		DeleteRule: nil,
		Schema: schema.NewSchema(
			&schema.SchemaField{
				Name:     "anime_no",
				Type:     schema.FieldTypeNumber,
				Required: true,
			},
			&schema.SchemaField{
				Name:     "subject",
				Type:     schema.FieldTypeText,
				Required: true,
			},
			&schema.SchemaField{
				Name:     "episode",
				Type:     schema.FieldTypeText,
				Required: true,
			},
			&schema.SchemaField{
				Name: "name",
				Type: schema.FieldTypeText,
			},
			&schema.SchemaField{
				Name: "website",
				Type: schema.FieldTypeText,
			},
		),
	}

	if err := app.Dao().SaveCollection(collection); err != nil {
		return err
	}

	return nil
}
