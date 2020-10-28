# Initial Set Up - Check Before You Start!

This is a very challenging to implement playground where we are using a new open source technology in order to be able to deliver it. As it is all very new and one of the first playgrounds to test this out please feel free to continue the playground from your local machine at any point. 

Do you have GO installed? 
If you are not using the playground provided infrastructure, please do the following:-
- Install GO
- Create a AWS account
- Set up the DynamoDB according to the cloudformation template at `dynamodb/dynamo.cf-template.yml`
- Set up a user that has permissions to CRUD items on the `flights` and `passengers` table
- Install AWS CLI on your local machine
- Create Access Key for the user in AWS and set them up for your AWS CLI on your local machine

# Stage 1: Setting up the project

If using the playground provided infrastructure:-

```
Go to in your browser `https://digital-meetup-signed-users.s3-eu-west-1.amazonaws.com/index.html`

Enter your meetup display name (this is case sensitive). 

A link to a Terminal (command line) and an IDE should appear

Please make a note of the animal you have been assigned from the link e.g. eagle
```

Go to the command line

If using the playground provided infrastructure:-

```
cd GraphQL
```

Create a directory for your project, and initialise it as a Go Module:

`mkdir flights`

`cd flights`

`go mod init flights`

Retrieve the required `gqlgen` package

`go get github.com/99designs/gqlgen`

Create the project skeleton

`go run github.com/99designs/gqlgen init`

Copy over a pre-prepared file using the command

`mkdir -p datalayer && cp ~/GraphQL/Introduction-to-GraphQL-with-GO/datalayer/datalayer.go datalayer/datalayer.go`

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

Run the command

`go run github.com/99designs/gqlgen generate`

Don't worry about the scary looking `validation failed` and `exit status 1` output from the command

Open and observe `graph/model/models_gen.go` it should contain a `Flight` and `Passenger` struct

Open the newly placed `datalayer/datalayer.go` and edit the file to replace `<YOUR_ANIMAL_NAME_HERE>` with your animal name

Open `graph/schema.resolvers.go`
Delete the content below and including `// !!! WARNING !!!`

Add the below into the file imports
```go
flights/datalayer
```

Find the func `Passengers` and replace the implementation with
```go
return datalayer.GetAllPassengers()
```

Run the command `go run ./server.go`

If you're using the provisioned infrastructure go to `http://your-animal.devopsplayground.org:8080`

If running locally then, in a web browser navigate to `http://localhost:8080`

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

NOTE: At this point there is no longer a use of `fmt.Errorf()` so please remove the unused `"fmt"` from the imports. This may come up multiple times below

Run the command `go run ./server.go`

If you're using the provisioned infrastructure go to `http://your-animal.devopsplayground.org:8080`

If running locally then, in a web browser navigate to `http://localhost:8080`

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

To do this open the file `dynamodb/flight_data.json` and replace `<YOUR_ANIMAL_NAME_HERE>` with your animal name

Then run the below command with the file:-

`aws dynamodb batch-write-item --region eu-west-2 --request-items file://~/GraphQL/Introduction-to-GraphQL-with-GO/dynamodb/flight_data.json`

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

If you're using the provisioned infrastructure go to `http://your-animal.devopsplayground.org:8080`

If running locally then, in a web browser navigate to `http://localhost:8080`

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

If you're using the provisioned infrastructure go to `http://your-animal.devopsplayground.org:8080`

If running locally then, in a web browser navigate to `http://localhost:8080`

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

Paste the below query into the left panel of the web page
```
query Flights {
  flights {
    number,
    passengers {
      name
    },
  }
}
```

Execute the query and you should see the result
```
{
  "data": {
    "flights": [
      {
        "number": "BA-386",
        "passengers": [
          {
            "name": "Bob"
          }
        ]
      },
      {
        "number": "BA-284",
        "passengers": null
      }
    ]
  }
}
```

The above result you might use for a mobile app as the screen is small so only a small number of details should be shown.
However if you were writing a desktop app instead then you may want to show more details.
You can easily change the query to return more details from the flights like so:-

```
query Flights {
  flights {
    number,
    passengers {
      name
    },
    capacity,
    captain
  }
}
```

For completeness, please modify the `Mutation` type to look like this:-
```
type Mutation {
  createPassenger(name: String!): Passenger!
  deletePassenger(passengerId: ID!): Boolean!
  bookFlight(flightNumber: String!, passengerId: ID!): Boolean!
  cancelBooking(flightNumber: String!, passengerId: ID!): Boolean!
}
```

Run the command `go run github.com/99designs/gqlgen generate`
and connect up the appropriate methods from the datalayer.
You can then have a play!