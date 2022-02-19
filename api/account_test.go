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

	"github.com/golang/mock/gomock"
	mockdb "github.com/gyu-young-park/simplebank/db/mock"
	db "github.com/gyu-young-park/simplebank/db/sqlc"
	"github.com/gyu-young-park/simplebank/util"
	"github.com/stretchr/testify/require"
)

type TestGetAccountAPISuite struct {
	name          string
	accountID     int64
	buildStubs    func(store *mockdb.MockStore)
	checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
}

func TestGetAccountAPI(t *testing.T) {
	account := randomAccount()

	okCase := TestGetAccountAPISuite{
		name:      "OK",
		accountID: account.ID,
		buildStubs: func(store *mockdb.MockStore) {
			store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
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
			server := NewServer(store)
			recorder := httptest.NewRecorder() // 진짜로 http server를 실행할 필요가 없으니 recorder만 넣는다.
			url := fmt.Sprintf("/accounts/%d", tc.accountID)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req) // 응답, 요청
			tc.checkResponse(t, recorder)
		})
	}
}

func randomAccount() db.Account {
	return db.Account{
		ID:       util.RandomInt(1, 1000),
		Owner:    util.RandomOwner(),
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
