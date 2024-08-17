package messaging

import (
	"time"

	"github.com/case-framework/case-backend/pkg/messaging/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (dbService *MessagingDBService) CreateSentSMSIndex(instanceID string) error {
	ctx, cancel := dbService.getContext()
	defer cancel()

	_, err := dbService.collectionSentSMS(instanceID).Indexes().CreateMany(
		ctx, []mongo.IndexModel{
			{
				Keys: bson.D{
					{Key: "userID", Value: 1},
					{Key: "sentAt", Value: 1},
					{Key: "messageType", Value: 1},
				},
			},
		},
	)

	return err
}

func (dbService *MessagingDBService) AddToSentSMS(instanceID string, sms types.SentSMS) (types.SentSMS, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	res, err := dbService.collectionSentSMS(instanceID).InsertOne(ctx, sms)
	if err != nil {
		return sms, err
	}
	sms.ID = res.InsertedID.(string)
	return sms, nil
}

func (dbService *MessagingDBService) CountSentSMSForUser(instanceID string, userID string, messageType string, sentAfter time.Time) (int64, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"userID": userID,
		"sentAt": bson.M{"$gt": sentAfter},
	}
	if messageType != "" {
		filter["messageType"] = messageType
	}

	return dbService.collectionSentSMS(instanceID).CountDocuments(ctx, filter)
}

func (dbService *MessagingDBService) GetAllSentSMSForUser(instanceID string, userID string, sentAfter time.Time) ([]types.SentSMS, error) {
	ctx, cancel := dbService.getContext()
	defer cancel()

	filter := bson.M{
		"userID": userID,
		"sentAt": bson.M{"$gt": sentAfter},
	}

	var sms []types.SentSMS
	cursor, err := dbService.collectionSentSMS(instanceID).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &sms); err != nil {
		return nil, err
	}
	return sms, nil
}
