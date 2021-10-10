package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var DB *gorm.DB
var err error

type Model struct {
	CreatedAt time.Time `json:"CreatedAt,omitempty"`
	UpdatedAt time.Time `json:"UpdatedAt,omitempty"`
}

type Customer struct {
	Model
	Id    int     `json:"id,omitempty"`
	Name  string  `json:"name,omitempty"`
	Total float64 `json:"total" sql:"type:float(11,2)"`
}

type Log struct {
	Model
	Id    int     `json:"id"`
	Cid   int     `json:"cid"`
	Type  string  `json:"type"`
	Total float64 `json:"total" sql:"type:float(11,2)"`
}

type Result struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Massage string      `json:"massage"`
}

var SumTotal float64

func main() {
	DB, err = gorm.Open("mysql", "root:someonep@ss@tcp(127.0.0.1:3306)/test_go?charset=utf8mb4&parseTime=True&loc=Local")

	if err != nil {
		log.Println("Connection Failed to Open")
	} else {
		log.Println("Connection Established")
	}

	DB.AutoMigrate(&Customer{})
	DB.AutoMigrate(&Log{})

	callhandlers()
}

func callhandlers() {

	myRoute := mux.NewRouter().StrictSlash(true)

	myRoute.HandleFunc("/create", create).Methods("POST")
	myRoute.HandleFunc("/income", income).Methods("POST")
	myRoute.HandleFunc("/withdraw", withdraw).Methods("POST")
	myRoute.HandleFunc("/getcustomertotal/{id}", getTotal).Methods("GET")
	myRoute.HandleFunc("/getlog/{id}", getlog).Methods("GET")

	log.Fatal(http.ListenAndServe(":8888", myRoute))
}

func create(w http.ResponseWriter, r *http.Request) {
	loadsRe, _ := ioutil.ReadAll(r.Body)

	var customer Customer
	json.Unmarshal(loadsRe, &customer)
	resultDB := DB.Create(&customer)

	var res Result
	if resultDB.RowsAffected > 0 {
		res = Result{Code: 200, Massage: "Success Create"}
	} else {
		res = Result{Code: 400, Data: customer, Massage: "Error Create"}
	}
	result, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func income(w http.ResponseWriter, r *http.Request) {
	re_custommerID, _ := strconv.Atoi(r.FormValue("id"))
	re_custommerTotal := r.FormValue("total")

	var customer_select Customer
	resultDB := DB.Where("id = ?", re_custommerID).Find(&customer_select)

	var res Result
	if resultDB.RowsAffected > 0 {
		custommerTotal, _ := strconv.ParseFloat(re_custommerTotal, 64)
		SumTotal = 0
		SumTotal = customer_select.Total + custommerTotal

		var customer Customer
		DB.Model(&customer).Where("id = ?", re_custommerID).Update("total", SumTotal)

		sqllog := Log{Cid: re_custommerID, Type: "income", Total: custommerTotal}
		DB.Create(&sqllog)

		res = Result{Code: 200, Data: customer, Massage: "Update Income to Customer Success"}
	} else {
		res = Result{Code: 400, Massage: "No Find Customer"}
	}

	result, err := json.Marshal(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func withdraw(w http.ResponseWriter, r *http.Request) {
	re_custommerID, _ := strconv.Atoi(r.FormValue("id"))
	re_custommerTotal := r.FormValue("total")

	var res Result
	var customer_select Customer
	resultDB := DB.Where("id = ?", re_custommerID).Find(&customer_select)

	if resultDB.RowsAffected > 0 {
		custommerTotal, _ := strconv.ParseFloat(re_custommerTotal, 64)
		SumTotal = 0

		var customer Customer
		if custommerTotal < customer_select.Total {
			SumTotal = customer_select.Total - custommerTotal
			DB.Model(&customer).Where("id = ?", re_custommerID).Update("total", SumTotal)

			sqllog := Log{Cid: re_custommerID, Type: "withdraw", Total: custommerTotal}
			DB.Create(&sqllog)

			res = Result{Code: 200, Data: customer, Massage: "Withdraw Success"}
		} else {
			res = Result{Code: 400, Massage: "No Withdraw"}
		}
	} else {
		res = Result{Code: 400, Massage: "No Withdraw"}
	}

	result, err := json.Marshal(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func getTotal(w http.ResponseWriter, r *http.Request) {
	Vars := mux.Vars(r)
	CustomerID, _ := strconv.Atoi(Vars["id"])

	var res Result
	var customer Customer
	resultDB := DB.Where("id = ?", CustomerID).First(&customer)

	if resultDB.RowsAffected > 0 {
		res = Result{Code: 200, Data: customer, Massage: "Get Success"}
	} else {
		res = Result{Code: 400, Data: customer, Massage: "No Find Customer"}
	}

	result, err := json.Marshal(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func getlog(w http.ResponseWriter, r *http.Request) {
	Vars := mux.Vars(r)
	CustomerID, _ := strconv.Atoi(Vars["id"])

	var res Result
	var log []Log
	resultDB := DB.Where("cid = ?", CustomerID).Find(&log)

	if resultDB.RowsAffected > 0 {
		res = Result{Code: 200, Data: log, Massage: "Get Success"}
	} else {
		res = Result{Code: 400, Data: log, Massage: "No Find Customer"}
	}

	result, err := json.Marshal(res)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}
