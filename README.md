# SkootServer

## Deployment instructions

1. Set up a postgresql database in the same ip as the server and port 5432 (localhost:5432) with the following tables:
```
CREATE TABLE booking (
    bookingid character varying(5) PRIMARY KEY,
    ammountpaid double precision NOT NULL,
    distancetravelled double precision NOT NULL,
    timetaken double precision NOT NULL,
    dateofbooking timestamp with time zone,
    email character varying(254) NOT NULL,
    scooterid character varying(5)
);
```

```
CREATE TABLE collector (
    collectorid character varying(5) PRIMARY KEY,
    iscollecting boolean NOT NULL,
    ischarging boolean NOT NULL,
    scooterscharged integer NOT NULL,
    ammountearned double precision NOT NULL,
    karmarating integer NOT NULL
);
```

```
CREATE TABLE rider (
    riderid character varying(5) PRIMARY KEY,
    fname character varying(15) NOT NULL,
    lname character varying(15) NOT NULL,
    email character varying(254) NOT NULL,
    password character varying(15) NOT NULL,
    iscollector boolean NOT NULL,
    creditcardno character varying(16),
    creditcardcvv integer
);
```

```
CREATE TABLE scooter (
    scooterid character varying(5) PRIMARY KEY,
    pose character varying(30) NOT NULL,
    posn character varying(30) NOT NULL,
    available boolean NOT NULL,
    battery character varying(3) NOT NULL,
    distancetravelled character varying(6)
);
```

```
CREATE TABLE staff (
    staffid character varying(5) PRIMARY KEY,
    fname character varying(15) NOT NULL,
    lname character varying(15) NOT NULL,
    email character varying(254) NOT NULL,
    address character varying(50) NOT NULL,
    dob date NOT NULL,
    teleno character varying(15) NOT NULL
);
```
2. Set up [Golang](https://golang.org/)
3. Run command ```go run *.go``` from the root directory of this repository
