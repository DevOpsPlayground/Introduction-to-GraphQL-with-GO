# Initial Set Up - Check Before You Start!

This is a very challenging to implement playground where we are using a new open source technology in order to be able to deliever it. As it is all very new and one of the first playgrounds to test this out please feel free to continue the playground from your local machine at any point. 

Do you have a GitHub account? If not sign up here: https://github.com/
You may want to create a throw away GitHub accont for this playground because of integrations and having to authorise GitHub for use on our instance. 
> Note: The instances will be destroyed shortly after the playground so any sensitive keys will be removed and not stored.

Do you have GO installed?
If you are not using the playground provided infrastructure, please do the following:-
- Install GO
- Create a AWS account
- Set up the DynamoDB according to the cloudformation template at `dynamodb/dynamo.cf-template.yml`
- Set up a user that has permissions to CRUD items on the `flights` and `passengers` table
- Install AWS CLI on your local machine
- Create Access Key for the user in AWS and set them up for your AWS CLI on your local machine

# Stage 1: Setting up the project

Create a directory for your project, and initialise it as a Go Module:

`mkdir flights`
`cd flights`
`go mod init github.com/<GITHUB_USERNAME>/flights`
`go get github.com/99designs/gqlgen`

Create the project skeleton

`go run github.com/99designs/gqlgen init`

# Stage 2: Creating the schema

Open `graph/schema.graphqls`
Delete the contents of this file and replace with

```go
type Flight {
  number: String!
  passengers: [Passenger!]
  capacity: Int!
  captain: String!
  plane: String!
}

type Passenger {
  id: ID!
  name: String!
}

type Query {
  passengers: [Passenger!]
}

type Mutation {
  createPassenger(name: String!): Passenger!
}

schema {
  query: Query
  mutation: Mutation
}
```

# Stage 3: Implementation!

Run the command `go run github.com/99designs/gqlgen generate`

Don't worry about the scary looking `validation failed` and `exit status 1` output from the command

Open and observe `graph/model/models_gen.go` it should contain a `Flight` and `Passenger` struct

Download `datalayer/datalayer.go` from this repository and place inside your `flights` project at the location `datalayer/datalayer.go`

Open `graph/schema.resolvers.go`
Delete the content below and including `// !!! WARNING !!!`

Find the func `Passengers` and replace the implementation with
```go
return datalayer.GetAllPassengers()
```

Run the command `go run ./server.go`

In a web browser navigate to `http://localhost:8080`

Paste the below query into the left panel of the web page
```
query Passengers {
  passengers {
    id,
    name
  }
}
```

Execute the query and you should see the result
```
{
  "data": {
    "passengers": null
  }
}
```

This is because the DynamoDB table doesn't have any data in it yet

Control-C the running command

Return to your `flights` project and open `graph/schema.resolvers.go`

Find the func `CreatePassenger` and replace the implementation with
```go
return datalayer.CreatePassenger(name)
```

Run the command `go run ./server.go`

Open the web browser to `http://localhost:8080` again

Paste the below query into the left panel of the web page
```
mutation CreatePassenger {
  createPassenger(name: "Bob") {
    id
  }
}
```

Execute the query and you should see the result
```
{
  "data": {
    "createPassenger": {
      "id": "<SOME_GUID_HERE>"
    }
  }
}
```

Now re-run `query Passengers` from above and you should see the id and name of the newly created passenger detailed

Control-C the running command

Now we've created a passenger lets put some data in the flights table

To do this run the below command using the json file from `dynamodb/flight_data.json`

`aws dynamodb batch-write-item --request-items file://flight_data.json`

Return to your `flights` project and open `graph/schema.graphqls`

Modify the `Query` type to look like this:-
```
type Query {
  flights: [Flight!]
  passengers: [Passenger!]
}
```

Run the command `go run github.com/99designs/gqlgen generate`

Open `graph/schema.resolvers.go`

Find the func `Flights` and replace the implementation with
```go
return datalayer.GetAllFlights()
```

Run the command `go run ./server.go`

Open the web browser to `http://localhost:8080` again

Paste the below query into the left panel of the web page
```
query Flights {
  flights {
    number
  }
}
```

Execute the query and you should see the result
```
{
  "data": {
    "flights": [
      {
        "number": "BA-386"
      },
      {
        "number": "BA-284"
      }
    ]
  }
}
```

Control-C the running command

Now lets book a passenger onto this flight

Return to your `flights` project and open `graph/schema.graphqls`

Modify the `Mutation` type to look like this:-
```
type Mutation {
  createPassenger(name: String!): Passenger!
  bookFlight(flightNumber: String!, passengerId: ID!): Boolean!
}
```

Run the command `go run github.com/99designs/gqlgen generate`

Open `graph/schema.resolvers.go`

Find the func `BookFlight` and replace the implementation with
```go
return datalayer.BookFlight(flightNumber, passengerID)
```

Run the command `go run ./server.go`

Open the web browser to `http://localhost:8080` again

Paste the below query into the left panel of the web page
```
mutation BookFlight {
  bookFlight(flightNumber: "BA-386", passengerId: "<GUID_OF_PASSENGER>")
}
```

Execute the query and you should see the result
```
{
  "data": {
    "bookFlight": true
  }
}
```