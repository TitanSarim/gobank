package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)



type APIServer struct {
	listenAddr string
	store      Storage
}

func NewApiServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{listenAddr: listenAddr, store: store}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/login", makeHTTPHandleFunc(s.handleLogin))

	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))

	router.HandleFunc("/account/{id}", withJwtAuth(makeHTTPHandleFunc(s.handleGetAccountByID), s.store))

	router.HandleFunc("/transfer", withJwtAuth(makeHTTPHandleFunc(s.handleTransferAccount), s.store))

	log.Println("Json API Server on port:", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) error{
	if r.Method != "POST" {
		return fmt.Errorf("method not supported")
	}

	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req) ; err != nil {
		return err
	}
	account, err := s.store.GetAccountByEmail(req.Email)
	
	if err != nil {
        return err
    }

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(req.Password)); err != nil {
    	return fmt.Errorf("invalid credentials")
	}

	tokenString, err := creatJWT(account)
	http.SetCookie(w, &http.Cookie{
		Name:     "x-jwt-token",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: false,   // Prevent JavaScript access
		Secure:   false,   // Set to true in production (requires HTTPS)
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Hour * 2),
	})
	if err != nil {
        return err
    }
	fmt.Println("JWT token: ", tokenString)
	return WriteJson(w, http.StatusOK, req)
}
 
func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET"{
		return s.handleGetAccounts(w, r)
	}
	if r.Method == "POST"{
		return s.handleCreateAccount(w, r)
	}
	

	return fmt.Errorf("methods not supported %s", r.Method)
}

func (s *APIServer) handleGetAccounts(w http.ResponseWriter, _ *http.Request) error {
	account, err := s.store.GetAccounts()

	if err != nil {
        return err
    }

	return WriteJson(w, http.StatusOK, account)
}

func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {

		id, err :=  getID(r)

		if err != nil {
			return err
		}

		account, err := s.store.GetAccountByID(id)

		if err != nil {
			return err
		}

		return WriteJson(w, http.StatusOK, account)
	} 

	if r.Method == "DELETE" {
		return s.handleDeleteAccount(w, r)
	}

	return fmt.Errorf("methods not supported %s", r.Method)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountReq); err != nil {
		return err
	}	

	account, err := NewAccount( createAccountReq.Email ,createAccountReq.FirstName, createAccountReq.LastName, createAccountReq.Password)
	if err != nil {
        return err
    }

	if err := s.store.CreateAccount(account); err != nil {
		return err
	}

	
	return WriteJson(w, http.StatusOK, account)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {

	id, err :=  getID(r)

	if err != nil {
        return err
    }

	if err := s.store.DeleteAccount(id); err != nil {
		return err
	}
	return WriteJson(w, http.StatusOK, map[string]int{"deleted": id})

}

func (s *APIServer) handleTransferAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("method not supported")
	}

	transferReq := new(TransferRequest)

	if err := json.NewDecoder(r.Body).Decode(transferReq); err != nil {
        return err
    }

	 // Extract accountNumber from cookie or headers
    var accountNumber int64
    cookie, err := r.Cookie("x-jwt-token")
	if err != nil {
		http.Error(w, "Token not found in cookies", http.StatusUnauthorized)
		return nil
	}
	tokenString := cookie.Value

	token, err := validateJwtAuth(tokenString)
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return nil
	}

    claims := token.Claims.(jwt.MapClaims)
 	accountNumber = int64(claims["accountNumber"].(float64))
	
    if accountNumber == 0 {
        return fmt.Errorf("accountNumber not found in cookies or headers")
    }

    transferReq.FromAccount = accountNumber
    trn, err := s.store.CreateTransfer(transferReq)
    if err != nil {
        return err
    }
    return WriteJson(w, http.StatusOK, trn)
}



func WriteJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

const jwtSecret = "as8d8aashd998aidsha89dakjasd98"

func creatJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expires": 15000,
		"id": account.ID,
		"accountNumber": account.Number,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func withJwtAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Extract token from cookie
        cookie, err := r.Cookie("x-jwt-token")
        if err != nil {
            http.Error(w, "Token not found in cookies", http.StatusUnauthorized)
            return
        }
        tokenString := cookie.Value

        token, err := validateJwtAuth(tokenString)
        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        claims := token.Claims.(jwt.MapClaims)
        userID := int(claims["id"].(float64))
		println(userID)
        // Validate user ID with database
        account, err := s.GetAccountByID(userID)
        if err != nil {
            http.Error(w, "Invalid account", http.StatusUnauthorized)
            return
        }

        // Validate account number from JWT claims
        if account.Number != int64(claims["accountNumber"].(float64)) {
            http.Error(w, "Invalid account", http.StatusUnauthorized)
            return
        }

        handlerFunc(w, r)
    }
}



func validateJwtAuth(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
	// Don't forget to validate the alg is what you expect:
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}

	// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
	return []byte(jwtSecret), nil
})
}

type apiFunc func (http.ResponseWriter, *http.Request)  error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil{
			// * handle error here
			WriteJson(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func getID(r *http.Request) (int, error) {
	idstr := mux.Vars(r)["id"]

    id, err := strconv.Atoi(idstr)
    if err != nil {
        return 0, fmt.Errorf("invalid id given by %s", idstr)
    }

    return id, nil
}