package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/julienschmidt/httprouter"
)

type ChaincodeStruct struct {
	Name string `json:"name"`
}
type CTorMsgStruct struct {
	Function string   `json:"function"`
	Args     []string `json:"args"`
}

type ParamStruct struct {
	Type          int             `json:"type"`
	ChainCode     ChaincodeStruct `json:"ChaincodeID"`
	CtorMsg       CTorMsgStruct   `json:"ctorMsg"`
	SecureContext string          `json:"secureContext"`
}

type Body struct {
	JsonRPc string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  ParamStruct `json:"params"`
	Id      int         `json:"id"`
}

var CHainCodeId string
var UserEntity string

func init() {
	CHainCodeId = "3aeb9793d67968f966f2b093c361c70cdbf7a2813a02f7a5da344386580d3b519899b73003b335c587e3d016d44b54eb7d8030bddddbc3e9abf05db81c20eaef"
	UserEntity = "admin"
}

func LoginHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) *HttpError {

	ctx := appengine.NewContext(r)

	username := r.FormValue("username")
	password := r.FormValue("pass")

	if username == "" || password == "" {
		failureLoginHandler(w, "username or password is empty", r)
		return nil
	}

	client := urlfetch.Client(ctx)

	var body Body
	body.JsonRPc = "2.0"
	body.Method = "query"
	body.Id = 0

	body.Params.Type = 1

	body.Params.ChainCode.Name = CHainCodeId //add chaincode Id here
	body.Params.SecureContext = UserEntity   //add value for valid user type 0

	body.Params.CtorMsg.Function = "read"
	body.Params.CtorMsg.Args = append(body.Params.CtorMsg.Args, username)

	j, aerr := json.Marshal(body)
	if aerr != nil {
		return NewErrorInternalServerError(aerr)
	}

	url := "https://74d34f81930a486d937d43cd423706a3-vp0.us.blockchain.ibm.com:5004/chaincode"

	air, err := http.NewRequest("POST", url, bytes.NewReader(j))
	air.Header.Set("Content-Type", "application/json")

	httpres, err := client.Do(air)

	if err != nil {
		failureLoginHandler(w, "error from chaincode"+err.Error(), r)
		return nil
	}

	datastring, _ := ioutil.ReadAll(httpres.Body)

	type ResultStruct struct {
		Message *string `json:"message"`
	}

	type outputfromChaincode struct {
		Result ResultStruct `json:"result"`
	}

	var oc outputfromChaincode

	err = json.Unmarshal(datastring, &oc)
	if err != nil {
		failureLoginHandler(w, "Invalid username or password", r)
		return nil
	}

	if oc.Result.Message == nil {
		failureLoginHandler(w, "Invalid username or password", r)
		return nil
	}

	expiration := time.Now().Add(365 * 24 * time.Hour)
	cookieUsername := http.Cookie{Name: "username", Value: username, Expires: expiration}

	http.SetCookie(w, &cookieUsername)

	type User struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		Balance  int    `json:"balance"`
	}

	var user User
	err = json.Unmarshal([]byte(*oc.Result.Message), &user)
	if err != nil {
		failureLoginHandler(w, "error while marshalling user object ", r)
		return nil
	}

	if user.Password != password {
		failureLoginHandler(w, "Invalid password", r)
		return nil
	}

	users, err := getUsers(r)
	if err != nil {
		return NewErrorInternalServerError(err)
	}

	t, _ := templates["user.html"]
	err = t.Execute(w, struct {
		Title        string
		Username     string
		Balance      string
		ErrorMessage string
		Users        []string
	}{
		"Wallet",
		user.Name,
		strconv.Itoa(user.Balance),
		"",
		users,
	})

	log.Infof(ctx, "cool2out")
	if err != nil {
		log.Errorf(ctx, err.Error())
		return NewErrorInternalServerError(err)
	}
	return nil
}

func getUserHomePage(username string, w http.ResponseWriter, r *http.Request) *HttpError {

	ctx := appengine.NewContext(r)

	client := urlfetch.Client(ctx)

	var body Body
	body.JsonRPc = "2.0"
	body.Method = "query"
	body.Id = 0

	body.Params.Type = 1

	body.Params.ChainCode.Name = CHainCodeId //add chaincode Id here
	body.Params.SecureContext = UserEntity   //add value for valid user type 0

	body.Params.CtorMsg.Function = "read"
	body.Params.CtorMsg.Args = append(body.Params.CtorMsg.Args, username)

	j, aerr := json.Marshal(body)
	if aerr != nil {
		return NewErrorInternalServerError(aerr)
	}

	url := "https://74d34f81930a486d937d43cd423706a3-vp0.us.blockchain.ibm.com:5004/chaincode"

	air, err := http.NewRequest("POST", url, bytes.NewReader(j))
	air.Header.Set("Content-Type", "application/json")

	httpres, err := client.Do(air)

	if err != nil {
		failureLoginHandler(w, "error from chaincode"+err.Error(), r)
		return nil
	}

	datastring, _ := ioutil.ReadAll(httpres.Body)

	type ResultStruct struct {
		Message *string `json:"message"`
	}

	type outputfromChaincode struct {
		Result ResultStruct `json:"result"`
	}

	var oc outputfromChaincode

	err = json.Unmarshal(datastring, &oc)
	if err != nil {
		failureLoginHandler(w, "Invalid username or password", r)
		return nil
	}

	if oc.Result.Message == nil {
		failureLoginHandler(w, "Invalid username or password", r)
		return nil
	}

	type User struct {
		Name     string `json:"name"`
		Password string `json:"password"`
		Balance  int    `json:"balance"`
	}

	var user User
	err = json.Unmarshal([]byte(*oc.Result.Message), &user)
	if err != nil {
		failureLoginHandler(w, "error while marshalling user object ", r)
		return nil
	}

	users, err := getUsers(r)
	if err != nil {
		return NewErrorInternalServerError(err)
	}

	t, _ := templates["user.html"]
	err = t.Execute(w, struct {
		Title        string
		Username     string
		Balance      string
		ErrorMessage string
		Users        []string
	}{
		"Wallet",
		user.Name,
		strconv.Itoa(user.Balance),
		"",
		users,
	})

	log.Infof(ctx, "cool2out")
	if err != nil {
		log.Errorf(ctx, err.Error())
		return NewErrorInternalServerError(err)
	}

	return nil
}

func createUserHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) *HttpError {

	ctx := appengine.NewContext(r)

	username := r.FormValue("username")
	password := r.FormValue("pass")
	balance := r.FormValue("balance")

	if username == "" || password == "" || balance == "" {
		failureUserCreationHandler(w, "username or password or balance is empty", r)
		return nil
	}

	fmt.Println(username, password, balance)

	client := urlfetch.Client(ctx)

	var body Body
	body.JsonRPc = "2.0"
	body.Method = "invoke"
	body.Id = 0

	body.Params.Type = 1

	body.Params.ChainCode.Name = CHainCodeId
	body.Params.SecureContext = "user_type1_0" //add value for valid user type 0

	body.Params.CtorMsg.Function = "create_user"
	body.Params.CtorMsg.Args = append(body.Params.CtorMsg.Args, username)
	body.Params.CtorMsg.Args = append(body.Params.CtorMsg.Args, password)
	body.Params.CtorMsg.Args = append(body.Params.CtorMsg.Args, balance)

	j, aerr := json.Marshal(body)
	if aerr != nil {
		return NewErrorInternalServerError(aerr)
	}

	url := "https://74d34f81930a486d937d43cd423706a3-vp0.us.blockchain.ibm.com:5004/chaincode"

	air, err := http.NewRequest("POST", url, bytes.NewReader(j))
	air.Header.Set("Content-Type", "application/json")

	httpres, err := client.Do(air)

	if err != nil {
		failureLoginHandler(w, "error from chaincode"+err.Error(), r)
		return nil
	}

	datastring, _ := ioutil.ReadAll(httpres.Body)

	type ResultStruct struct {
		Message *string `json:"message"`
	}

	type outputfromChaincode struct {
		Result ResultStruct `json:"result"`
	}

	var oc outputfromChaincode

	err = json.Unmarshal(datastring, &oc)
	if err != nil {
		failureUserCreationHandler(w, "Invalid username or password", r)
		return nil
	}

	if oc.Result.Message == nil {
		failureUserCreationHandler(w, "failed creation at cc", r)
		return nil
	}

	failureUserCreationHandler(w, "Successfully cretaed user.", r)

	return nil
}

func failureLoginHandler(w http.ResponseWriter, msg string, r *http.Request) {

	t, _ := templates["home.html"]
	err := t.Execute(w, struct {
		Title        string
		ErrorMessage string
	}{
		"Login",
		msg,
	})
	if err != nil {
		fmt.Fprint(w, err.Error())
	}

}

func failureUserCreationHandler(w http.ResponseWriter, msg string, r *http.Request) {

	t, _ := templates["admin.html"]

	users, err := getUsers(r)
	if err != nil {
		return
	}

	err = t.Execute(w, struct {
		Title        string
		Users        []string
		ErrorMessage string
	}{
		"Admin",
		users,
		msg,
	})
	if err != nil {
		fmt.Fprint(w, err.Error())
	}

}

func HomeHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) *HttpError {

	ctx := appengine.NewContext(r)

	username, err := r.Cookie("username")
	if err != nil {

		t, _ := templates["login.html"]
		err := t.Execute(w, nil)
		if err != nil {
			return NewErrorInternalServerError(err)
		}
		return nil
	} else {
		log.Infof(ctx, "username present"+username.Value)
		return getUserHomePage(username.Value, w, r)
	}

}

func getUsers(r *http.Request) ([]string, error) {

	ctx := appengine.NewContext(r)

	client := urlfetch.Client(ctx)

	var body Body
	body.JsonRPc = "2.0"
	body.Method = "query"
	body.Id = 0

	body.Params.Type = 1

	body.Params.ChainCode.Name = 3aeb9793d67968f966f2b093c361c70cdbf7a2813a02f7a5da344386580d3b519899b73003b335c587e3d016d44b54eb7d8030bddddbc3e9abf05db81c20eaef //add chaincode Id here
	body.Params.SecureContext = admin   //add value for valid user type 0

	body.Params.CtorMsg.Function = "read"
	body.Params.CtorMsg.Args = append(body.Params.CtorMsg.Args, "users")

	j, aerr := json.Marshal(body)
	if aerr != nil {
		return nil, aerr
	}

	url := "https://74d34f81930a486d937d43cd423706a3-vp0.us.blockchain.ibm.com:5004/chaincode"

	air, err := http.NewRequest("POST", url, bytes.NewReader(j))
	air.Header.Set("Content-Type", "application/json")

	httpres, err := client.Do(air)

	if err != nil {
		return nil, err
	}

	datastring, _ := ioutil.ReadAll(httpres.Body)

	type ResultStruct struct {
		Message *string `json:"message"`
	}

	type outputfromChaincode struct {
		Result ResultStruct `json:"result"`
	}

	var oc outputfromChaincode

	err = json.Unmarshal(datastring, &oc)
	if err != nil {
		return nil, err
	}

	if oc.Result.Message == nil {
		return nil, err
	}

	var users []string

	err = json.Unmarshal([]byte(*oc.Result.Message), &users)

	if err != nil {
		return nil, err
	}

	return users, nil

}

func AdminHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) *HttpError {
	t, _ := templates["admin.html"]

	users, err := getUsers(r)
	if err != nil {
		return NewErrorInternalServerError(err)
	}

	err = t.Execute(w, struct {
		Title        string
		Users        []string
		ErrorMessage string
	}{
		"Admin",
		users,
		"",
	})
	if err != nil {
		return NewErrorInternalServerError(err)
	}
	return nil
}

func transferMoney(firstUser string, secondUser string, amount string, r *http.Request) error {

	ctx := appengine.NewContext(r)
	client := urlfetch.Client(ctx)

	var body Body
	body.JsonRPc = "2.0"
	body.Method = "invoke"
	body.Id = 0

	body.Params.Type = 1

	body.Params.ChainCode.Name =3aeb9793d67968f966f2b093c361c70cdbf7a2813a02f7a5da344386580d3b519899b73003b335c587e3d016d44b54eb7d8030bddddbc3e9abf05db81c20eaef
	body.Params.SecureContext = admin //add value for valid user type 0

	body.Params.CtorMsg.Function = "transaction"
	body.Params.CtorMsg.Args = append(body.Params.CtorMsg.Args, firstUser)
	body.Params.CtorMsg.Args = append(body.Params.CtorMsg.Args, secondUser)
	body.Params.CtorMsg.Args = append(body.Params.CtorMsg.Args, amount)

	j, aerr := json.Marshal(body)
	if aerr != nil {
		return aerr
	}

	url := "https://74d34f81930a486d937d43cd423706a3-vp0.us.blockchain.ibm.com:5004/chaincode"

	air, err := http.NewRequest("POST", url, bytes.NewReader(j))
	air.Header.Set("Content-Type", "application/json")

	httpres, err := client.Do(air)

	if err != nil {
		return err
	}
	datastring, _ := ioutil.ReadAll(httpres.Body)

	type ResultStruct struct {
		Message *string `json:"message"`
	}

	type outputfromChaincode struct {
		Result ResultStruct `json:"result"`
	}

	var oc outputfromChaincode

	err = json.Unmarshal(datastring, &oc)
	if err != nil {
		return err
	}

	if oc.Result.Message == nil {
		return err
	}

	return nil

}

func TransfermoneyHandler(w http.ResponseWriter, r *http.Request, p httprouter.Params) *HttpError {
	ctx := appengine.NewContext(r)

	username, err := r.Cookie("username")
	if err != nil {
		fmt.Fprint(w, "cookie for original user not found")
		return nil
	}

	if username == nil {
		fmt.Fprint(w, "cookie for original user not found")
		return nil
	}

	secondUser := r.FormValue("selected_user")
	amount := r.FormValue("amount")

	err = transferMoney(username.Value, secondUser, amount, r)
	if err != nil {
		log.Errorf(ctx, err.Error())
		return NewErrorInternalServerError(err)
	}

	fmt.Fprint(w, "<html>Successfully Transferred <a href=/user>Go Back</a>  <html>")

	return nil
}
