package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const minSecretKeySize = 32

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) (TokenMaker, error) {
	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeySize)
	}
	return &JWTMaker{secretKey}, nil
}

// 순서는 이렇게된다. 먼저 NewWithClaims로 헤더와 페이로드 부분을 넣은 구조체를 만들고 SignedString호출하는데 내부적으로 SigningString를 호출하여 헤더와 페이로드를 마샬링한 후에
// 인코딩한다. 그 다음 key로 서명한 뒤 서명이 완료된다.
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return jwtToken.SignedString([]byte(maker.secretKey))
}

func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// keyFunc으로 token을 검증을 하고, secret 키를 반환하면 이를 기반으로 검증하게된다.. 가령 create한 토큰 알고리즘과 다른 지 확인하는 경우가 있다.
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC) // SigningMethodHS256은 SigningMethodHMAC의 인스턴스이다. 변환에 실패하면 ok가 false로 나온다.
		if !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	}
	// 두번째 인자로로 페이로드의 포인터를 전송하면, 이를 token에서 claim으로 받아 토큰의 페이로드 값을 디코딩한다. 그래서 아래에 jwtToken.Claims.(*Payload)로 접근하는 것이다.
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		verr, ok := err.(*jwt.ValidationError)
		if ok && errors.Is(verr.Inner, ErrExpiredToken) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}
	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}
	return payload, nil
}
