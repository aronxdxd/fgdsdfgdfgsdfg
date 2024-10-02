package funcs

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"

	"project/db/operations"
	"project/global"
)

type TopUser struct {
	Place     int    `json:"place"`
	FirstName string `json:"first_name"`
	Balance   int32  `json:"balance"`
}

type TopResult struct {
	UserPosition TopUser   `json:"user_position"`
	Top10        []TopUser `json:"top_10"`
}

func HandleTop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		global.RespondWithError(w, http.StatusMethodNotAllowed, "Only POST method is allowed")
		return
	}

	var userInfo global.UserInfo
	if err := json.NewDecoder(r.Body).Decode(&userInfo); err != nil {
		global.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	users, err := operations.GetAllUsers(r.Context())
	if err != nil {
		global.RespondWithError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	sortUsersByBalance(users)

	result := processUsers(users, userInfo.TelegramID)

	global.RespondWithJson(w, http.StatusOK, map[string]interface{}{
		"result": result,
	})
}

func sortUsersByBalance(users []map[string]interface{}) {
	sort.Slice(users, func(i, j int) bool {
		balanceI, okI := users[i]["balance"].(int32)
		balanceJ, okJ := users[j]["balance"].(int32)
		return okI && okJ && balanceI > balanceJ
	})
}

func processUsers(users []map[string]interface{}, targetUserID int) TopResult {
	var result TopResult

	if len(users) >= 10 {
		users = users[:10]
	}
	
	result.Top10 = make([]TopUser, 0, 10)

	for i, user := range users {
		firstName := user["first_name"].(string)
		balance := user["balance"].(int32)
		position := i + 1

		topUser := TopUser{
			Place:     position,
			FirstName: firstName,
			Balance:   balance,
		}
		log.Println(user["user_id"], targetUserID)
		if fmt.Sprint(user["user_id"]) == strconv.Itoa(targetUserID) {
			log.Println(topUser)
			result.UserPosition = topUser
		}

		result.Top10 = append(result.Top10, topUser)
	}

	return result
}
