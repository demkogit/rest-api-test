package apiserver

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"

	"gihub.com/demkogit/rest-api/internal/app/store"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type APIServer struct {
	config *Config
	logger *logrus.Logger
	router *mux.Router
	store  *store.Store
}

func New(config *Config) *APIServer {
	return &APIServer{
		config: config,
		logger: logrus.New(),
		router: mux.NewRouter(),
	}
}

func (s *APIServer) Start() error {
	if err := s.configureLogger(); err != nil {
		return err
	}

	s.configureRouter()

	if err := s.configureStore(); err != nil {
		return err
	}

	s.logger.Info("starting api server")

	return http.ListenAndServe(s.config.BindAddr, s.router)
}

func (s *APIServer) configureLogger() error {
	level, err := logrus.ParseLevel(s.config.LogLevel)
	if err != nil {
		return err
	}

	s.logger.SetLevel(level)

	return nil
}

func (s *APIServer) configureRouter() {
	s.router.HandleFunc("/hello", s.handleHello())
	s.router.HandleFunc("/auth/{guid}", s.authHandler()).Methods("GET")
	s.router.HandleFunc("/refresh/", s.refreshHandler()).Methods("POST")
}

func (s *APIServer) configureStore() error {
	st := store.New(s.config.Store)
	if err := st.Open(); err != nil {
		return err
	}

	s.store = st

	return nil
}

func (s *APIServer) handleHello() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		io.WriteString(rw, "Hello")
	}
}

type User struct {
	GUID         primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Nickname     string             `json:"nickname,omitempty" bson:"nickname,omitempty"`
	RefreshToken string             `json:"refreshToken,omitempty" bson:"refreshToken,omitempty"`
	AccessToken  string             `json:"accessToken,omitempty" bson:"accessToken,omitempty"`
	Expire       primitive.DateTime `json:"expire,omitempty" bson:"expire,omitempty"`
}

func (s *APIServer) authHandler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("content-type", "application/json")
		param := mux.Vars(r)

		u, err := s.store.User().FindById(param["guid"])
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}

		newRefreshToken := generateRefreshToken()
		oldRefreshToken := u.RefreshToken

		accessToken, err := generateAccessToken()

		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}

		if err := s.store.User().UpdateRefreshToken(oldRefreshToken, newRefreshToken); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}
		u.RefreshToken = newRefreshToken
		u.AccessToken = accessToken

		json.NewEncoder(rw).Encode(u)
	}
}

var tokenEncodeString = "dasdqwr12rq1!23%%dd"

func generateAccessToken() (string, error) {
	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.StandardClaims{
			ExpiresAt: jwt.TimeFunc().Add(600000).Unix(),
			IssuedAt:  jwt.TimeFunc().Unix(),
		},
	)
	return token.SignedString([]byte(tokenEncodeString))
}

func generateRefreshToken() string {
	password := []byte("Password123")
	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	b64HashPass := base64.StdEncoding.EncodeToString(hashedPassword)
	return b64HashPass
}

func (s *APIServer) refreshHandler() http.HandlerFunc {

	type request struct {
		RefreshToken string `json:"refreshToken"`
	}

	return func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("content-type", "application/json")

		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte(`{ "message asda": "` + err.Error() + `" }`))
			return
		}

		if req.RefreshToken == "" {
			return
		}

		u, err := s.store.User().FindByRefreshToken(req.RefreshToken)
		if err != nil {
			rw.WriteHeader(http.StatusUnauthorized)
			rw.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}

		newRefreshToken := generateRefreshToken()

		accessToken, err := generateAccessToken()
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}

		if err := s.store.User().UpdateRefreshToken(req.RefreshToken, newRefreshToken); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(`{ "message": "` + err.Error() + `" }`))
			return
		}

		u.RefreshToken = newRefreshToken
		u.AccessToken = accessToken

		json.NewEncoder(rw).Encode(u)
	}
}
