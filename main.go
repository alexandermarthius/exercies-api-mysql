package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alexandermarthius/exercies-api-mysql/config"
	"github.com/alexandermarthius/exercies-api-mysql/models"
	"github.com/alexandermarthius/exercies-api-mysql/user"
	"github.com/alexandermarthius/exercies-api-mysql/utils"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type M map[string]interface{}

type MyClaims struct {
	jwt.StandardClaims
	UserID   int    `json:"id"`
	Username string `json:"username"`
	Active   int    `json:"active"`
	RoleID   int    `json:"role_name"`
}

var APPLICATION_NAME = "Exercies API MYSQL"
var LOGIN_EXPIRATION_DURATION = time.Duration(1) * time.Minute * 10
var JWT_SIGNING_METHOD = jwt.SigningMethodHS256
var JWT_SIGNATURE_KEY = []byte("aaaaaaaaaa")

func main() {
	db, e := config.MySQL()

	if e != nil {
		log.Fatal(e)
	}

	if ep := db.Ping(); ep != nil {
		log.Fatal(ep)
	}
	fmt.Println("berhasil connect ke MySQL")

	mux := new(utils.CustomMux)
	mux.RegisterMiddleware(MiddlewareJWTAuthorization)

	mux.HandleFunc("/login", HandlerLogin)
	mux.HandleFunc("/index", HandlerIndex)

	mux.HandleFunc("/user", GetUser)
	mux.HandleFunc("/user/info", GetUserInfo)
	mux.HandleFunc("/user/create", CreateUser)
	mux.HandleFunc("/user/update", UpdateUser)
	mux.HandleFunc("/user/delete", DeleteUser)

	server := new(http.Server)
	server.Handler = mux
	server.Addr = ":6060"

	fmt.Println("Starting server at", server.Addr)
	server.ListenAndServe()
}

func HandlerIndex(w http.ResponseWriter, r *http.Request) {
	userInfo := r.Context().Value("userInfo").(jwt.MapClaims)
	message := fmt.Sprintf("hello %s (%s)", userInfo["Username"], userInfo["Group"])
	w.Write([]byte(message))
}

func HandlerLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Unsupported http method", http.StatusBadRequest)
		return
	}

	username, password, ok := r.BasicAuth()
	fmt.Println(username, password)
	if !ok {
		http.Error(w, "Invalid username or password", http.StatusBadRequest)
		return
	}

	ok, userInfo := authenticateUser(w, username, password)
	if !ok {
		http.Error(w, "Invalid username or password", http.StatusBadRequest)
		return
	}

	claims := MyClaims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    APPLICATION_NAME,
			ExpiresAt: time.Now().Add(LOGIN_EXPIRATION_DURATION).Unix(),
		},
		UserID:   userInfo.ID,
		Username: userInfo.Username,
		Active:   userInfo.Active,
		RoleID:   userInfo.RoleID,
	}

	token := jwt.NewWithClaims(
		JWT_SIGNING_METHOD,
		claims,
	)

	signedToken, err := token.SignedString(JWT_SIGNATURE_KEY)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tokenString, _ := json.Marshal(M{"token": signedToken})
	w.Write([]byte(tokenString))
}

func authenticateUser(w http.ResponseWriter, username string, password string) (bool, models.User) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dtUser, err := user.GetByUsername(ctx, username)
	if err != nil {
		log.Fatal(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(dtUser.Password), []byte(password))
	if err != nil {
		return false, models.User{}
	}

	return true, dtUser
}

func MiddlewareJWTAuthorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}

		tokenString, err := getToken(w, r)
		if err != nil {
			utils.ResponseJSON(w, err, http.StatusBadRequest)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Signing method invalid")
			} else if method != JWT_SIGNING_METHOD {
				return nil, fmt.Errorf("Signing method invalid")
			}

			return JWT_SIGNATURE_KEY, nil
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(context.Background(), "userInfo", claims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func GetUserInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		getToken, err := getToken(w, r)
		if err != nil {
			utils.ResponseJSON(w, err, http.StatusBadRequest)
			return
		}
		userInfo, err := extractClaims(getToken)
		if err != nil {
			utils.ResponseJSON(w, err, http.StatusBadRequest)
			return
		}

		/*
			// Belum berhasil assert interface ke int
			id, ok := userInfo["id"].(int)
			if !ok {
				log.Fatal("internal error")
			}
		*/
		users, err := user.GetOne(ctx, userInfo["id"])

		if err != nil {
			log.Fatal(err)
		}

		utils.ResponseJSON(w, users, http.StatusOK)
		return
	}
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ok, err := isAdmin(w, r)
		if err != nil {
			utils.ResponseJSON(w, err, http.StatusBadRequest)
			return
		}

		if !ok {
			utils.ResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		users, err := user.GetAll(ctx)

		if err != nil {
			log.Fatal(err)
		}

		utils.ResponseJSON(w, users, http.StatusOK)
		return
	}
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		ok, err := isAdmin(w, r)
		if err != nil {
			utils.ResponseJSON(w, err, http.StatusBadRequest)
			return
		}

		if !ok {
			utils.ResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Gunakan Content-Type: application/json", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var usermodel models.User

		if err := json.NewDecoder(r.Body).Decode(&usermodel); err != nil {
			utils.ResponseJSON(w, err, http.StatusBadRequest)
			return
		}

		if err := user.Insert(ctx, usermodel); err != nil {
			utils.ResponseJSON(w, err, http.StatusInternalServerError)
			return
		}

		res := map[string]string{
			"status": "succesfully",
		}

		utils.ResponseJSON(w, res, http.StatusCreated)
		return
	}

	http.Error(w, "hai, jangan macam2!", http.StatusMethodNotAllowed)
	return
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPut {
		ok, err := isAdmin(w, r)
		if err != nil {
			utils.ResponseJSON(w, err, http.StatusBadRequest)
			return
		}

		if !ok {
			utils.ResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Gunakan Content-Type: application/json", http.StatusBadRequest)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var usermodel models.User

		if err := json.NewDecoder(r.Body).Decode(&usermodel); err != nil {
			utils.ResponseJSON(w, err, http.StatusBadRequest)
			return
		}

		if err := user.Update(ctx, usermodel); err != nil {
			utils.ResponseJSON(w, err, http.StatusInternalServerError)
			return
		}

		res := map[string]string{
			"status": "successfully",
		}

		utils.ResponseJSON(w, res, http.StatusOK)
		return
	}

	utils.ResponseJSON(w, "jangan macam2!", http.StatusMethodNotAllowed)
	return
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		ok, err := isAdmin(w, r)
		if err != nil {
			utils.ResponseJSON(w, err, http.StatusBadRequest)
			return
		}

		if !ok {
			utils.ResponseJSON(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var usermodel models.User

		id := r.URL.Query().Get("id")

		if id == "" {
			utils.ResponseJSON(w, "id tidak boleh kosong", http.StatusBadRequest)
			return
		}

		usermodel.ID, _ = strconv.Atoi(id)

		if err := user.Delete(ctx, usermodel); err != nil {
			kesalahan := map[string]string{
				"error": fmt.Sprintf("%v", err),
			}

			utils.ResponseJSON(w, kesalahan, http.StatusInternalServerError)
			return
		}

		res := map[string]string{
			"status": "succesfully",
		}

		utils.ResponseJSON(w, res, http.StatusOK)
		return
	}

	utils.ResponseJSON(w, "jangan macam2!", http.StatusMethodNotAllowed)
	return
}

func isAdmin(w http.ResponseWriter, r *http.Request) (bool, error) {
	getToken, err := getToken(w, r)
	if err != nil {
		utils.ResponseJSON(w, err, http.StatusBadRequest)
		return false, errors.New("Invalid token")
	}
	userInfo, err := extractClaims(getToken)
	if err != nil {
		utils.ResponseJSON(w, err, http.StatusBadRequest)
		return false, errors.New("Invalid token")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	users, err := user.GetOne(ctx, userInfo["id"])

	if users.RoleID == 1 {
		return true, nil
	} else {
		return false, nil
	}
}

func getToken(w http.ResponseWriter, r *http.Request) (string, error) {
	authorizationHeader := r.Header.Get("Authorization")
	if !strings.Contains(authorizationHeader, "Bearer") {
		return "", errors.New("Invalid token")
	}

	tokenString := strings.Replace(authorizationHeader, "Bearer ", "", -1)

	return tokenString, nil
}

func extractClaims(tokenStr string) (jwt.MapClaims, error) {
	hmacSecretString := JWT_SIGNATURE_KEY // Value
	hmacSecret := []byte(hmacSecretString)
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// check token signing method etc
		return hmacSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("Invalid token")
	}
}
