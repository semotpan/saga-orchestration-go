#!/bin/sh

# httpie
echo
echo "Place a room reservation successfully"
http POST http://localhost:8080/api/v1/reservations < room-reservation-placement.json

echo "Place a room reservation with invalid payment, room should be released (compensation)"
http POST http://localhost:8080/api/v1/reservations < invalid-payment.json

echo "Place a room reservation with unavailable room"
http POST http://localhost:8080/api/v1/reservations < invalid-room-taken.json

# curl
#echo
#echo "Place a room reservation successfully"
#curl -i -X POST -H Accept:application/json -H Content-Type:application/json http://localhost:8080/api/v1/reservations -d @room-reservation-placement.json

#echo
#echo "Place a room reservation with invalid payment, room should be released (compensation)"
#curl -i -X POST -H Accept:application/json -H Content-Type:application/json http://localhost:8080/api/v1/reservations -d @invalid-payment.json

#echo
#echo "Place a room reservation with unavailable room"
#curl -i -X POST -H Accept:application/json -H Content-Type:application/json http://localhost:8080/api/v1/reservations -d @invalid-room-taken.json
