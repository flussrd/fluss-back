package repository

import (
	"context"
	"errors"
	"time"

	"github.com/flussrd/fluss-back/app/river-management/models"
	repository "github.com/flussrd/fluss-back/app/river-management/repositories/modules"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	databaseName = "flussDB"

	mongoDuplicateCode = 11000
)

type mongoRepository struct {
	client *mongo.Client
}

// New returns a new mongoRepository instance
func New(client *mongo.Client) repository.Repository {
	return mongoRepository{client: client}
}

func (r mongoRepository) GetModule(ctx context.Context, moduleID string) (models.Module, error) {
	modulesCollection := r.getModulesCollection()

	module := models.Module{}

	docID, err := primitive.ObjectIDFromHex(moduleID)

	if err != nil {
		return models.Module{}, err
	}

	err = modulesCollection.FindOne(ctx, bson.M{
		"_id": docID,
	}).Decode(&module)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.Module{}, repository.ErrNotFound
	}

	if err != nil {
		return models.Module{}, err
	}

	return module, nil
}

func (r mongoRepository) GetModuleByPhoneNumber(ctx context.Context, phoneNumber string) (models.Module, error) {
	modulesCollection := r.getModulesCollection()

	module := models.Module{}

	err := modulesCollection.FindOne(ctx, bson.M{
		"phoneNumber": phoneNumber,
	}).Decode(&module)

	if errors.Is(err, mongo.ErrNoDocuments) {
		return models.Module{}, repository.ErrNotFound
	}

	if err != nil {
		return models.Module{}, err
	}

	return module, nil
}

func (r mongoRepository) GetAllModulesWithOutPagination(ctx context.Context) ([]models.Module, error) {
	return r.getModulesWithoutPagination(ctx, bson.M{})
}

func (r mongoRepository) GetAllModules(ctx context.Context) ([]models.Module, string, error) {
	return nil, "", nil
}

func (r mongoRepository) GetModulesByRiverWithoutPagination(ctx context.Context, riverID string) ([]models.Module, error) {
	return r.getModulesWithoutPagination(ctx, bson.M{"riverID": riverID})
}

func (r mongoRepository) getModulesWithoutPagination(ctx context.Context, filter bson.M) ([]models.Module, error) {
	modulesCollection := r.getModulesCollection()

	modules := []models.Module{}

	result, err := modulesCollection.Find(ctx, filter)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, repository.ErrNotFound
	}

	err = result.All(ctx, &modules)
	if err != nil {
		return nil, err
	}

	return modules, nil
}

func (r mongoRepository) GetModulesByRiver(ctx context.Context) ([]models.Module, string, error) {
	return nil, "", nil
}

func (r mongoRepository) SaveModule(ctx context.Context, module models.Module) (models.Module, error) {
	collection := r.getModulesCollection()

	module.CreationDate = time.Now()
	module.UpdateDate = time.Now()

	_, err := collection.InsertOne(ctx, module)
	if errors.As(err, &mongo.WriteException{}) {
		mongoErr, _ := err.(mongo.WriteException)
		switch mongoErr.WriteErrors[0].Code {
		case mongoDuplicateCode:
			return models.Module{}, repository.ErrDuplicateFields
		}
	}

	if err != nil {
		return models.Module{}, err
	}

	return module, nil
}

func (r mongoRepository) UpdateModule(ctx context.Context, moduleID string, options models.ModuleUpdateOptions) error {
	collection := r.getModulesCollection()

	updateMap := map[string]interface{}{}
	if options.State != "" {
		updateMap["state"] = options.State
	}

	if options.Location.Lat != 0 && options.Location.Lng != 0 {
		updateMap["location"] = map[string]interface{}{
			"latitude":  options.Location.Lat,
			"longitude": options.Location.Lng,
		}
	}

	updateResult, err := collection.UpdateOne(ctx, bson.M{
		"_id": moduleID,
	}, bson.M{
		"$set": updateMap,
	})

	if updateResult != nil && updateResult.MatchedCount == 0 {
		return repository.ErrNotFound
	}

	if err != nil {
		return err
	}

	return nil
}

func (r mongoRepository) getModulesCollection() *mongo.Collection {
	return r.client.Database(databaseName).Collection("modules")
}
