package datalayer

import (
	"flights/graph/model"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

const AnimalName = "<YOUR_ANIMAL_NAME_HERE>"

var PassengersTableName string = fmt.Sprintf("playground-passengers-%s", AnimalName)

var FlightsTableName string = fmt.Sprintf("playground-flights-%s", AnimalName)

func initialiseDb() *dynamodb.DynamoDB {
	// Initialize a session that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials
	// and region from the shared configuration file ~/.aws/config.
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	return dynamodb.New(sess)
}

func scanTable(svc *dynamodb.DynamoDB, tableName string) (*dynamodb.ScanOutput, error) {
	// Build the query input parameters
	params := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	}

	// Make the DynamoDB Query API call
	return svc.Scan(params)
}

func CreatePassenger(name string) (*model.Passenger, error) {
	svc := initialiseDb()

	item := model.Passenger{
		ID:   uuid.New().String(),
		Name: name,
	}

	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		fmt.Println("Got error marshalling new passenger item:")
		fmt.Println(err.Error())
		return nil, err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(PassengersTableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		fmt.Println("Got error calling PutItem:")
		fmt.Println(err.Error())
		return nil, err
	}

	return &item, nil
}

func DeletePassenger(passengerId string) (bool, error) {
	svc := initialiseDb()

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(passengerId),
			},
		},
		TableName: aws.String(PassengersTableName),
	}

	_, err := svc.DeleteItem(input)
	if err != nil {
		fmt.Println("Got error calling DeleteItem")
		fmt.Println(err.Error())
		return false, err
	}

	return true, nil
}

// Adds "setItem" to the StringSet (SS) identified by "setAttribute" on the record with a
// a partition key of "keyAttribute" with the value of "key" in the Dynamo table "table".
func addToSet(db *dynamodb.DynamoDB, table, keyAttribute, key, setAttribute, setItem string) error {
	_, err := db.UpdateItem(&dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#0": &setAttribute,
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":0": {SS: []*string{&setItem}},
		},
		Key: map[string]*dynamodb.AttributeValue{
			keyAttribute: {S: &key},
		},
		TableName:        &table,
		UpdateExpression: aws.String("ADD #0 :0"),
	})
	return err
}

// Deletes "setItem" from the StringSet (SS) identified by "setAttribute" on the record with a
// a partition key of "keyAttribute" with the value of "key" in the Dynamo table "table".
func deleteFromSet(db *dynamodb.DynamoDB, table, keyAttribute, key, setAttribute, setItem string) error {
	_, err := db.UpdateItem(&dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#0": &setAttribute,
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":0": {SS: []*string{&setItem}},
		},
		Key: map[string]*dynamodb.AttributeValue{
			keyAttribute: {S: &key},
		},
		TableName:        &table,
		UpdateExpression: aws.String("DELETE #0 :0"),
	})
	return err
}

func BookFlight(flightNumber string, passengerId string) (bool, error) {
	svc := initialiseDb()

	err := addToSet(svc, FlightsTableName, "number", flightNumber, "passengers", passengerId)

	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}

	return true, nil
}

func CancelBooking(flightNumber string, passengerId string) (bool, error) {
	svc := initialiseDb()

	err := deleteFromSet(svc, FlightsTableName, "number", flightNumber, "passengers", passengerId)

	if err != nil {
		fmt.Println(err.Error())
		return false, err
	}

	return true, nil
}

type DynamoFlight struct {
	Number     string
	Passengers []string
	Capacity   int
	Captain    string
	Plane      string
}

func GetAllFlights() ([]*model.Flight, error) {
	svc := initialiseDb()

	result, err := scanTable(svc, FlightsTableName)

	if err != nil {
		fmt.Println("Query API call failed:")
		fmt.Println((err.Error()))
		return nil, err
	}

	var flights []*model.Flight

	for _, dynamoItem := range result.Items {
		item := DynamoFlight{}

		err = dynamodbattribute.UnmarshalMap(dynamoItem, &item)

		if err != nil {
			fmt.Println("Got error unmarshalling:")
			fmt.Println(err.Error())
			return nil, err
		}

		flight, err := convertDynamoFlightToFlight(item)

		if err != nil {
			return nil, err
		}

		flights = append(flights, flight)
	}

	return flights, nil
}

func convertDynamoFlightToFlight(dynamoFlight DynamoFlight) (*model.Flight, error) {
	flight := model.Flight{
		Number:   dynamoFlight.Number,
		Capacity: dynamoFlight.Capacity,
		Captain:  dynamoFlight.Captain,
		Plane:    dynamoFlight.Plane,
	}

	for _, passengerId := range dynamoFlight.Passengers {
		passenger, err := GetPassenger(passengerId)

		if err != nil {
			fmt.Println("Query API call failed:")
			fmt.Println((err.Error()))
			return nil, err
		}

		flight.Passengers = append(flight.Passengers, passenger)
	}

	return &flight, nil
}

func GetPassenger(passengerId string) (*model.Passenger, error) {
	svc := initialiseDb()

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(passengerId),
			},
		},
		TableName: aws.String(PassengersTableName),
	}

	dynamoItem, err := svc.GetItem(input)

	if err != nil {
		fmt.Println("Query API call failed:")
		fmt.Println((err.Error()))
		return nil, err
	}

	item := model.Passenger{}

	err = dynamodbattribute.UnmarshalMap(dynamoItem.Item, &item)

	if err != nil {
		fmt.Println("Got error unmarshalling:")
		fmt.Println(err.Error())
		return nil, err
	}

	return &item, nil
}

func GetAllPassengers() ([]*model.Passenger, error) {
	svc := initialiseDb()

	result, err := scanTable(svc, PassengersTableName)

	if err != nil {
		fmt.Println("Query API call failed:")
		fmt.Println((err.Error()))
		return nil, err
	}

	var passengers []*model.Passenger

	for _, dynamoItem := range result.Items {
		item := model.Passenger{}

		err = dynamodbattribute.UnmarshalMap(dynamoItem, &item)

		if err != nil {
			fmt.Println("Got error unmarshalling:")
			fmt.Println(err.Error())
			return nil, err
		}

		passengers = append(passengers, &item)
	}

	return passengers, nil
}
