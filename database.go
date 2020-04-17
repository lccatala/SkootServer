package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

const (
	dbhost = "localhost"
	dbport = 5432
	dbuser = "alpasfly"
	dbname = "skoot"
)

var users map[string]string
var bookingids map[string]string

var db sql.DB
var psqlInfo string

func initDB() {
	psqlInfo = fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		dbhost, dbport, dbuser, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	CheckError(err)
	defer db.Close() // TODO maybe we should close on shutdown

	err = db.Ping()
	CheckError(err)

	loadUsers()

	LogInfo("Successfully connected to database")
}

func loadUsers() {
	users = make(map[string]string)
	bookingids = make(map[string]string)
	db, err := sql.Open("postgres", psqlInfo)
	CheckError(err)
	err = db.Ping()
	rows, err := db.Query("SELECT email, password FROM rider;")
	CheckError(err)
	defer rows.Close()
	for rows.Next() {
		var email string
		var password string
		err = rows.Scan(&email, &password)
		CheckError(err)
		users[email] = password
	}

	err = rows.Err()
	CheckError(err)
}

// Runs an INSERT, UPDATE or DELETE statement
func runStatement(sqlStatement string) {
	db, err := sql.Open("postgres", psqlInfo)
	CheckError(err)
	err = db.Ping()
	_, err = db.Exec(sqlStatement)
	CheckError(err)
	err = db.Close()
	CheckError(err)
}

func getRows(sqlStatement string) (rows *sql.Rows) {
	db, err := sql.Open("postgres", psqlInfo)
	CheckError(err)
	err = db.Ping()
	CheckError(err)
	rows, err = db.Query(sqlStatement)
	CheckError(err)
	return
}

func addUser(creds credentials) bool {

	// Check if user exists in cache
	_, exists := users[creds.Email]
	if exists {
		LogInfo("Tried to sign up user " + creds.Email + ", which already exists")
		return false
	}

	// Insert user in DB
	runStatement(`INSERT INTO 
	rider (riderid, fname, lname, email, password, iscollector, creditcardno, creditcardcvv)
	VALUES (` + strconv.Itoa(len(users)) + `, 
			'` + creds.Fname + `', '` + creds.Lname + `', '` + creds.Email + `', 
			'` + creds.Password + `', false, '` + creds.CreditCardNo + `', '` + creds.CVV + `');`)

	users[creds.Email] = creds.Password
	LogInfo("Signed up user " + creds.Email)
	return true
}

func authUser(email string, password string) (authorized bool) {
	expectedPassword, exists := users[email]

	if !exists {
		authorized = false
		return
	}

	// Reading the password from 'users' map
	if expectedPassword != password {
		LogInfo("User " + email + " failed to authenticate")
		authorized = false
		return
	}

	LogInfo("Authorizing user " + email)
	authorized = true
	return
}

func loginUser(creds credentials) (resp response) {
	resp.Authorized = authUser(creds.Email, creds.Password)
	if !resp.Authorized {
		return
	}

	rows := getRows("SELECT fname, lname, creditcardno, creditcardcvv FROM rider where email = '" + creds.Email + "'")
	defer rows.Close()
	for rows.Next() {
		var fname string
		var lname string
		var creditcardno string
		var creditcardcvv string

		err := rows.Scan(&fname, &lname, &creditcardno, &creditcardcvv)
		CheckError(err)

		resp.Fname = fname
		resp.Lname = lname
		resp.CreditCardNo = creditcardno
		resp.CVV = creditcardcvv
	}

	resp.Authorized = true
	return
}

func rentScooter(email string, code string) bool {
	available := scooterIsAvailable(code)
	if available {
		// Make scooter unavailable
		runStatement(`UPDATE scooter
		SET available = false
		WHERE scooterid = '` + code + `';`)

		var count int
		rows := getRows("SELECT COUNT(*) FROM booking")
		defer rows.Close()
		for rows.Next() {
			err := rows.Scan(&count)
			CheckError(err)
		}

		bookingids[email] = strconv.Itoa(count)

		// Create booking
		runStatement(`INSERT INTO booking
		VALUES ('` + strconv.Itoa(count) + `', 0, 0, 0, NOW(), '` + email + `', '` + code + `');`)
	}
	return available
}

func stopScooterRent(email string) {
	// Get scooter id
	var id string
	idrows := getRows("SELECT scooterid FROM booking WHERE bookingid = '" + bookingids[email] + "';")
	defer idrows.Close()
	for idrows.Next() {
		err := idrows.Scan(&id)
		CheckError(err)
	}

	// Make scooter available
	runStatement("UPDATE scooter SET available = true WHERE scooterid = '" + id + "';")

	// Update timetaken in booking table
	runStatement(`UPDATE booking 
	SET timetaken = EXTRACT(EPOCH FROM NOW() - dateofbooking) / 60
	WHERE scooterid = '` + id + `' and email = '` + email + `';`)

	// Calculate ride cost
	var distance float32
	var time float32
	costrows := getRows("SELECT distancetravelled, timetaken FROM booking WHERE bookingid = '" + bookingids[email] + "';")
	defer costrows.Close()
	for costrows.Next() {
		err := costrows.Scan(&distance, &time)
		CheckError(err)
	}
	cost := distance*3 + time*4

	// TODO: Update user wallet. This will require integration with PayPal or Stripe

	// Update booking table
	runStatement(`UPDATE booking 
	SET ammountpaid = ` + fmt.Sprintf("%f", cost) + `
	WHERE bookingid = '` + bookingids[email] + `'`)

	bookingids[email] = ""
}

func getUserRides(email string) (r rides) {
	rows := getRows("SELECT dateofbooking, timetaken FROM booking WHERE email = '" + email + "';")
	defer rows.Close()
	var startTime, timeTaken string
	i := 0
	for rows.Next() && i < 5 {
		rows.Scan(&startTime, &timeTaken)
		r.Rides[i] = "Time: " + startTime + "\n Length: " + timeTaken + " minutes."
		i++
	}
	return
}

func changeEmail(oldEmail string, newEmail string) (resp response) {
	runStatement("UPDATE rider SET email = '" + newEmail + "' WHERE email = '" + oldEmail + "';")
	runStatement("UPDATE booking SET email = '" + newEmail + "' WHERE email = '" + oldEmail + "';")

	users[newEmail] = users[oldEmail]
	delete(users, oldEmail)

	bookingids[newEmail] = bookingids[oldEmail]
	delete(bookingids, oldEmail)

	resp = getUserName(newEmail)
	return
}

func changePassword(email string, password string) (resp response) {
	runStatement("UPDATE rider SET password = '" + password + "' WHERE email = '" + email + "';")
	resp = getUserName(email)
	users[email] = password
	return
}

func changeName(email string, name string) (resp response) {
	names := strings.Split(name, "|")
	runStatement("UPDATE rider SET fname = '" + names[0] + "', lname = '" + names[1] + "' WHERE email = '" + email + "';")
	resp = getUserName(email)
	return
}

func changeCreditCard(email string, data string) (resp response) {
	nums := strings.Split(data, "|")
	runStatement("UPDATE rider SET creditcardno = '" + nums[0] + "', creditcardcvv = '" + nums[1] + "' WHERE email = '" + email + "';")
	resp = getUserName(email)
	return
}

func getRiderID(email string) (riderid string) {
	rows := getRows("SELECT riderid FROM rider WHERE email = '" + email + "';")
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&riderid)
		CheckError(err)
	}
	return
}

func writeApplication(id string, email string, letter string) {
	f, err := os.Create("applications/" + id + ".txt")
	defer f.Close()
	CheckError(err)
	f.WriteString("id: " + id + "\n\n" + "email: " + email + "\n\n" + letter + "\n")
	CheckError(err)
	f.Sync()
}

// Write application to text file or remove themselves directly from collectors
func changeCollector(email string, data string) (resp response) {
	comps := strings.Split(data, "|@|")
	riderid := getRiderID(email)

	if comps[0] == "true" {
		LogInfo("Registered application of user " + email + " for becomming a collector")
		writeApplication(riderid, email, comps[1])
	} else {
		LogInfo("User " + email + " withdrawed from being a collector")
		runStatement("DELETE FROM collector WHERE collectorid = '" + riderid + "';")
		runStatement("UPDATE rider SET iscollector = false WHERE riderid = '" + riderid + "';")
	}
	resp = getUserName(email)
	return
}

func getUserName(email string) (resp response) {
	rows := getRows("SELECT fname, lname FROM rider WHERE email = '" + email + "';")
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&resp.Fname, &resp.Lname)
		CheckError(err)
	}
	return
}

func changeSettings(sr settingRequest) (resp response) {
	switch sr.Setting {
	case "Email":
		resp = changeEmail(sr.Email, sr.Value)
	case "Password":
		resp = changePassword(sr.Email, sr.Value)
	case "Name":
		resp = changeName(sr.Email, sr.Value)
	case "CreditCard":
		resp = changeCreditCard(sr.Email, sr.Value)
	case "Collector":
		resp = changeCollector(sr.Email, sr.Value)
	}
	resp.Authorized = true
	return
}

func getCollector(sr settingRequest) (resp response) {
	rows := getRows("SELECT iscollector, riderid FROM rider WHERE email = '" + sr.Email + "';")
	defer rows.Close()
	var isCol = false
	var riderid string
	for rows.Next() {
		err := rows.Scan(&isCol, &riderid)
		CheckError(err)
	}
	if isCol {
		resp.CVV = "true"
		collectingRows := getRows("SELECT iscollecting FROM collector WHERE collectorid = '" + riderid + "';")
		defer collectingRows.Close()
		for collectingRows.Next() {
			err := collectingRows.Scan(&resp.CreditCardNo)
			CheckError(err)
		}
	} else {
		resp.CVV = "false"
	}
	return
}

func scooterIsAvailable(code string) (available bool) {
	rows := getRows("SELECT available FROM scooter WHERE scooterid = '" + code + "'")
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&available)
		CheckError(err)
	}
	return
}

func collectScooter(creds credentials) (resp response) {
	if !scooterIsAvailable(creds.Data) {
		resp.CVV = "unavailable"
		return
	}

	resp.CVV = "available"
	// Make scooter unavailable
	runStatement(`UPDATE scooter
	SET available = false
	WHERE scooterid = '` + creds.Data + `';`)

	// Set collector as "iscollecting"
	runStatement(`UPDATE collector
	SET iscollecting = true, ischarging = true
	WHERE collectorid = '` + getRiderID(creds.Email) + `'`)

	return
}

func returnCollectedScooter(creds credentials) (resp response) {
	if scooterIsAvailable(creds.Data) {
		resp.CVV = "available"
		return
	}

	resp.CVV = "unavailable"

	// Mark scooter available
	runStatement(`UPDATE scooter
	SET available = true
	WHERE scooterid = '` + creds.Data + `';`)

	// Set collector as "is not collecting"
	runStatement(`UPDATE collector
	SET iscollecting = false, ischarging = false
	WHERE collectorid = '` + getRiderID(creds.Email) + `'`)

	return
}
