package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mockdb "github.com/gyu-young-park/simplebank/db/mock"
	db "github.com/gyu-young-park/simplebank/db/sqlc"
	"github.com/gyu-young-park/simplebank/util"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}

	err := util.CheckPassword(e.password, arg.HashedPassword)
	if err != nil {
		return false
	}
	e.arg.HashedPassword = arg.HashedPassword
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

type TestAPISuite struct {
	name          string
	body          gin.H
	buildStubs    func(store *mockdb.MockStore)
	checkResponse func(recorder *httptest.ResponseRecorder)
}

func TestCreateUserAPI(t *testing.T) {
	user, password := randomUser(t)

	okCase := TestAPISuite{
		name: "OK",
		body: gin.H{
			"username":  user.Username,
			"password":  password,
			"full_name": user.FullName,
			"email":     user.Email,
		},
		buildStubs: func(store *mockdb.MockStore) {
			arg := db.CreateUserParams{
				Username: user.Username,
				FullName: user.FullName,
				Email:    user.Email,
			}
			store.EXPECT().
				CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
				Times(1).
				Return(user, nil)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusOK, recorder.Code)
			requireBodyMatchUser(t, recorder.Body, user)
		},
	}

	internalError := TestAPISuite{
		name: "InternalError",
		body: gin.H{
			"username":  user.Username,
			"password":  password,
			"full_name": user.FullName,
			"email":     user.Email,
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				CreateUser(gomock.Any(), gomock.Any()).
				Times(1).
				Return(user, sql.ErrConnDone)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusInternalServerError, recorder.Code)
		},
	}

	duplicateUsername := TestAPISuite{
		name: "DuplicateUsername",
		body: gin.H{
			"username":  user.Username,
			"password":  password,
			"full_name": user.FullName,
			"email":     user.Email,
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				CreateUser(gomock.Any(), gomock.Any()).
				Times(1).
				Return(user, &pq.Error{Code: "23505"})
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusForbidden, recorder.Code)
		},
	}

	invalidUsername := TestAPISuite{
		name: "InvalidUsername",
		body: gin.H{
			"username":  "invalid-user#1",
			"password":  password,
			"full_name": user.FullName,
			"email":     user.Email,
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				CreateUser(gomock.Any(), gomock.Any()).
				Times(0)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusBadRequest, recorder.Code)
		},
	}

	invalidEmail := TestAPISuite{
		name: "InvalidEmail",
		body: gin.H{
			"username":  user.Username,
			"password":  password,
			"full_name": user.FullName,
			"email":     "invalid-email",
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				CreateUser(gomock.Any(), gomock.Any()).
				Times(0)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusBadRequest, recorder.Code)
		},
	}

	tooShortPassword := TestAPISuite{
		name: "TooShortPassword",
		body: gin.H{
			"username":  user.Username,
			"password":  "123",
			"full_name": user.FullName,
			"email":     "invalid-email",
		},
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().
				CreateUser(gomock.Any(), gomock.Any()).
				Times(0)
		},
		checkResponse: func(recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusBadRequest, recorder.Code)
		},
	}

	testCase := []TestAPISuite{okCase,
		internalError,
		duplicateUsername,
		invalidUsername,
		invalidEmail,
		tooShortPassword,
	}

	for i := range testCase {
		tc := testCase[i]
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			store := mockdb.NewMockStore(mockController)
			// build stub₩
			tc.buildStubs(store)

			//start test server and send request
			server := NewServer(store)
			recorder := httptest.NewRecorder() // 진짜로 http server를 실행할 필요가 없으니 recorder만 넣는다.

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/users"
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req) // 응답, 요청
			tc.checkResponse(recorder)
		})
	}
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}

func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotUser db.User
	err = json.Unmarshal(data, &gotUser)

	require.NoError(t, err)
	require.Equal(t, user.Username, gotUser.Username)
	require.Equal(t, user.FullName, gotUser.FullName)
	require.Equal(t, user.Email, gotUser.Email)
	require.Empty(t, gotUser.HashedPassword)
}
