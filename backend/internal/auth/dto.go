package auth

import (
	"github.com/google/uuid"

	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/httpx"
	"github.com/rusu-cristian/miraj-impex-v3/backend/internal/platform/timex"
)

// Wire types of SPEC §7.1.

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// Validate: an issue on each empty field (answer Q4).
func (r loginRequest) Validate(v *httpx.Validator) {
	loginFields{login: NormalizeLogin(r.Login), password: r.Password}.Validate(v)
}

type userDTO struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Login string    `json:"login"`
	Role  Role      `json:"role"`
}

type farmDTO struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Type FarmType  `json:"type"`
}

// loginResponse is 200 { user, farms, today } (D13: the /me item shapes).
type loginResponse struct {
	User  userDTO   `json:"user"`
	Farms []farmDTO `json:"farms"`
	Today timex.Day `json:"today"`
}

type meResponse struct {
	User      userDTO         `json:"user"`
	Farms     []farmDTO       `json:"farms"`
	Today     timex.Day       `json:"today"`
	Now       timex.LocalTime `json:"now"`
	StartPage string          `json:"startPage"`
}

// startPage is where every role lands after sign-in (answer Q7: S2 Acasă, D121).
const startPage = "/acasa"

func toUserDTO(u User) userDTO {
	return userDTO(u)
}

func toFarmDTOs(farms []Farm) []farmDTO {
	out := make([]farmDTO, 0, len(farms))
	for _, f := range farms {
		out = append(out, farmDTO(f))
	}
	return out
}
