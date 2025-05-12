# Parking lot system

The application built using Go 1.23 and consist 5 API's (the detail of what and how to use the API will be explain later below). There are multiple floors in the parking lot.
Each floor has parking spots for different vehicle types: bicycles (B), motorcycles (M) and automobiles (A).
Parking spots are arranged in rows and columns. Some of the parking spots are inactive will be mark as 'X'.


# How to Run
Run this command in your terminal
```
make gorun
```

# API's

## Set Parking Lot Map
The request will receive CSV file contain map of the parking lot. The structure of this CSV file is floor, row, column number and vehicle type (B, M, A. X mean inactive). Here are the example of the file
|floor|row|column|vehicle_type|
|-----|---|------|------------|
|1    |1  |1     |B           |
|1    |1  |2     |M           |
|1    |1  |3     |A           |
|1    |1  |4     |X           |
|1    |1  |5     |X           |
|1    |2  |1     |B           |
|1    |2  |2     |M           |
|1    |2  |3     |A           |
|1    |2  |4     |X           |
|1    |2  |5     |X           |
|2    |1  |1     |B           |
|2    |1  |2     |M           |
|2    |1  |3     |A           |
|2    |1  |4     |X           |
|2    |1  |5     |X           |
|2    |2  |1     |B           |
|2    |2  |2     |M           |
|2    |2  |3     |A           |
|2    |2  |4     |X           |
|2    |2  |5     |X           |

Request
```
curl --location 'http://localhost:8080/v1/vehicle/parking-map' \
--form 'file=@"/Users/bayuanggara/projects/Test/parking-lot-map.csv"'
```
Response
```
HTTP 200 OK
```

## Park Vehicle
Given a vehicle type, assign an empty parking spot id and map the vehicleNumber. spot_id is floor-row-column. If no free spot is found, it will return an error.

Request
```
curl --location 'http://localhost:8080/v1/vehicle/park' \
--header 'Content-Type: application/json' \
--data '{
    "vehicle_type": "A",
    "vehicle_number": "B1234AB"
}'
```
Response
```json
{
    "spot_id": "1-2-3"
}
```

## Unpark Vehicle
Removes vehicle from parking spot. It will return an error for failure to unpark a vehicle.

Request
```
curl --location --request DELETE 'http://localhost:8080/v1/vehicle/park?spot_id=1-2-3&vehicle_number=B1234AB'
```
Response
```
HTTP 200 OK
```

## Available Spot
Display the every free spots for specific vehicle type

Request
```
curl --location 'http://localhost:8080/v1/vehicle/available-spot?vehicle_type=B'
```
Response
```json
{
    "spot_ids": [
        "1-1-1",
        "1-2-1",
        "1-3-1",
        "1-4-1",
        "1-5-1",
        "2-1-1",
        "2-2-1",
        "2-3-1",
        "2-4-1",
        "2-5-1",
        "3-1-1",
        "3-2-1",
        "3-3-1",
        "3-4-1",
        "3-5-1",
        "4-1-1",
        "4-2-1",
        "4-3-1",
        "4-4-1",
        "4-5-1",
        "5-1-1",
        "5-2-1",
        "5-3-1",
        "5-4-1",
        "5-5-1"
    ]
}
```

## Search Vehicle
Search a vehicle parking spot by vehicle number. If the vehicle has been unparked, it will return its last spot_id.

Request
```
curl --location 'http://localhost:8080/v1/vehicle/search?vehicle_number=B1234AB'
```
Response
```json
{
    "spot_id": "1-2-3"
}
```

# Tech Stack
  - Database -> [SQLite](https://www.sqlite.org)
  - Redis -> [EchoVault/SugarDB](https://github.com/EchoVault/SugarDB)