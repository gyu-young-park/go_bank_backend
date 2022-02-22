package token

import "time"

// 다양한 토큰 메이커를 제공하여 여러 알고리즘을 사용하는 토큰을 사용하도록 한다.
type TokenMaker interface {
	CreateToken(username string, duration time.Duration) (string, error)
	VerifyToken(token string) (*Payload, error)
}
