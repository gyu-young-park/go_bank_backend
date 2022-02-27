package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mockdb "github.com/gyu-young-park/simplebank/db/mock"
	db "github.com/gyu-young-park/simplebank/db/sqlc"
	"github.com/gyu-young-park/simplebank/token"
	"github.com/gyu-young-park/simplebank/util"
	"github.com/stretchr/testify/require"
)

type TestGetAccountAPISuite struct {
	name          string
	accountID     int64
	setAuth       func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker)
	buildStubs    func(store *mockdb.MockStore)
	checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
}

func TestGetAccountAPI(t *testing.T) {
	user, _ := randomUser(t)
	account := randomAccount(user.Username)

	okCase := TestGetAccountAPISuite{
		name:      "OK",
		accountID: account.ID,
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
		},
		setAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
			addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
		checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusOK, recorder.Code)
			requreBodyMatchAccount(t, recorder.Body, account)
		},
	}

	notFoundCase := TestGetAccountAPISuite{
		name:      "Not Found",
		accountID: account.ID,
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, sql.ErrNoRows)
		},
		setAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
			addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
		checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusNotFound, recorder.Code)
		},
	}

	connectionErrCase := TestGetAccountAPISuite{
		name:      "Internal Error",
		accountID: account.ID,
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, sql.ErrConnDone)
		},
		setAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
			addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
		checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusInternalServerError, recorder.Code)
		},
	}

	invalidIdCase := TestGetAccountAPISuite{
		name:      "Invalid Id",
		accountID: 0,
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().GetAccount(gomock.Any(), gomock.Any()).Times(0)
		},
		setAuth: func(t *testing.T, request *http.Request, tokenMaker token.TokenMaker) {
			addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.Username, time.Minute)
		},
		checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusBadRequest, recorder.Code)
		},
	}

	testCase := []TestGetAccountAPISuite{okCase, notFoundCase, connectionErrCase, invalidIdCase}

	for i := range testCase {
		tc := testCase[i]
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			store := mockdb.NewMockStore(mockController)
			// build stub₩
			tc.buildStubs(store)

			//start test server and send request
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder() // 진짜로 http server를 실행할 필요가 없으니 recorder만 넣는다.

			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setAuth(t, req, server.tokenMaker)
			server.router.ServeHTTP(recorder, req) // 응답, 요청
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount(owner string) db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    owner,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func requreBodyMatchAccount(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var gotAccount db.Account
	err = json.Unmarshal(data, &gotAccount)
	require.NoError(t, err)
	require.Equal(t, account, gotAccount)
}
