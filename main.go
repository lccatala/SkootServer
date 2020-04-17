package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
)

type credentials struct {
	Fname        string
	Lname        string
	Email        string
	Password     string
	CreditCardNo string
	CVV          string
	Phone        string
	Data         string
}

type response struct {
	Authorized   bool
	Fname        string
	Lname        string
	Email        string
	Password     string
	CreditCardNo string
	CVV          string
	Phone        string
}

type rides struct {
	Authorized bool
	Rides      [5]string
}

type settingRequest struct {
	Email    string
	Password string
	Setting  string
	Value    string
}

func CheckError(e error) {
	if e != nil {
		panic(e)
	}
}

func readSettingRequest(req *http.Request) (sr settingRequest) {
	body, err := ioutil.ReadAll(req.Body)
	CheckError(err)

	err = json.Unmarshal(body, &sr)
	CheckError(err)
	return
}

func readCredentials(req *http.Request) (creds credentials) {
	body, err := ioutil.ReadAll(req.Body)
	CheckError(err)

	err = json.Unmarshal(body, &creds)
	CheckError(err)
	return
}

func respond(w io.Writer, resp response) {
	LogTrace("Responding with auth " + strconv.FormatBool(resp.Authorized))
	JSONResp, err := json.Marshal(&resp)
	CheckError(err)
	w.Write(JSONResp)
}

func respondRides(w io.Writer, resp rides) {
	JSONResp, err := json.Marshal(&resp)
	CheckError(err)
	w.Write(JSONResp)
}

func loginHandler(w http.ResponseWriter, req *http.Request) {
	creds := readCredentials(req)
	response := loginUser(creds)
	respond(w, response)
}

func signupHandler(w http.ResponseWriter, req *http.Request) {
	creds := readCredentials(req)
	response := response{}
	response.Authorized = addUser(creds)
	respond(w, response)
}

func getBookingHandler(w http.ResponseWriter, req *http.Request) {
	creds := readCredentials(req)
	response := loginUser(creds)
	if response.Authorized {
		response.CVV = bookingids[creds.Email]
	}
	respond(w, response)
}

func rentHandler(w http.ResponseWriter, req *http.Request) {
	creds := readCredentials(req)
	LogTrace("User " + creds.Email + " renting scooter " + creds.Data)
	response := loginUser(creds)
	response.CVV = "unavailable" // Store availability of scooter in CVV
	if response.Authorized && rentScooter(creds.Email, creds.Data) {
		response.CVV = "available"
	}
	respond(w, response)
}

func stopRentHandler(w http.ResponseWriter, req *http.Request) {
	creds := readCredentials(req)
	LogTrace("User " + creds.Email + " stopping rental of scooter")
	response := loginUser(creds)
	if response.Authorized {
		stopScooterRent(creds.Email)
	}
	respond(w, response)
}

func ridesHandler(w http.ResponseWriter, req *http.Request) {
	creds := readCredentials(req)
	var r rides
	if authUser(creds.Email, creds.Password) {
		r = getUserRides(creds.Email)
		r.Authorized = true
	}
	respondRides(w, r)
}

func settingsHandler(w http.ResponseWriter, req *http.Request) {
	sr := readSettingRequest(req)
	var resp response
	if authUser(sr.Email, sr.Password) {
		resp = changeSettings(sr)
	}
	respond(w, resp)
}

func getCollectorHandler(w http.ResponseWriter, req *http.Request) {
	sr := readSettingRequest(req)
	var resp response
	if authUser(sr.Email, sr.Password) {
		resp = getCollector(sr)
		resp.Authorized = true
	}
	respond(w, resp)
}

func collectHandler(w http.ResponseWriter, req *http.Request) {
	creds := readCredentials(req)
	var resp response
	if authUser(creds.Email, creds.Password) {
		resp = collectScooter(creds)
		resp.Authorized = true
	}
	respond(w, resp)
}

func returnCollectedHandler(w http.ResponseWriter, req *http.Request) {
	creds := readCredentials(req)
	var resp response
	if authUser(creds.Email, creds.Password) {
		resp = returnCollectedScooter(creds)
		resp.Authorized = true
	}
	respond(w, resp)
}

func indexHandler(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	CheckError(err)
}

func main() {
	LogInfo("Initializing server...")
	initDB()
	http.HandleFunc("/index", indexHandler)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/rent", rentHandler)
	http.HandleFunc("/stopRent", stopRentHandler)
	http.HandleFunc("/getBooking", getBookingHandler)
	http.HandleFunc("/rides", ridesHandler)
	http.HandleFunc("/settings", settingsHandler)
	http.HandleFunc("/getCollector", getCollectorHandler)
	http.HandleFunc("/collect", collectHandler)
	http.HandleFunc("/returnCollected", returnCollectedHandler)
	LogInfo("Done!")
	err := http.ListenAndServe("localhost:8080", nil)
	CheckError(err)
}
