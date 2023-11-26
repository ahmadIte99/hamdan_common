package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Subscription struct {
	Microservices []string      `json:"microservices"`
	Note          []interface{} `json:"note"`
	Id            string        `json:"_id"`
	Name          string        `json:"name"`
	Storage       string        `json:"storage"`
	StartDate     time.Time     `json:"startDate"`
	EndDate       time.Time     `json:"endDate"`
	Status        string        `json:"status"`
	Trash         bool          `json:"trash"`
	CreatedBy     string        `json:"createdBy"`
	UpdatedBy     string        `json:"updatedBy"`
	CreatedAt     string        `json:"createdAt"`
	UpdatedAt     string        `json:"updatedAt"`
}

type Pagination struct {
	TotalPages float64 `json:"totalPages"`
	PerPage    int64   `json:"perPage"`
	TotalCount int64   `json:"totalCount"`
}

type Subscriptions struct {
	Contents   []Subscription `json:"contents"`
	Pagination Pagination     `json:"pagination"`
}

type ClientSubscription struct {
	SubscriptionData map[string]Subscription
}

var ClientSubscriptionInstance = &ClientSubscription{
	SubscriptionData: map[string]Subscription{},
}

func getClientIdFromUrl(url string) string {
	var r string

	r = strings.ReplaceAll(url, "/", "-")
	r = strings.ReplaceAll(r, "\\", "-")
	j := strings.Split(r, "-")
	if len(j) > 2 {
		return j[2]
	}
	return ""

}

func (cs *ClientSubscription) CheckSubscription(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println(r.URL.RequestURI())
		var Client string
		headerParams := ExtractHeaderParams(r)
		if headerParams.Client != "" {
			Client = headerParams.Client
		} else if clientId := r.URL.Query().Get("clientId"); clientId != "" {
			Client = clientId
		} else {
			Client = getClientIdFromUrl(r.URL.RequestURI())
		}
		if subscription, found := cs.SubscriptionData[Client]; found {

			//check for subscription date
			now := time.Now()
			then := subscription.EndDate
			if then.After(now) {
				next.ServeHTTP(w, r)
			} else {
				http.Error(w, "Customer subscription expired", http.StatusBadRequest)
			}

		} else {
			var reqOpt *RequestParams = &RequestParams{
				Service: "manager",
				Path:    "subscriptions",
				Method:  "GET",
				Data:    map[string]interface{}{},
				Header:  headerParams,
			}
			resp, err := CallService(reqOpt)
			fmt.Println("subscriptions error: ", err)

			if err != nil {
				http.Error(w, "manager service down!", http.StatusInternalServerError)
			} else if resp.StatusCode != 200 {
				http.Error(w, "manager service not available", http.StatusBadRequest)
			} else {
				var res Subscriptions
				decodeErr := json.NewDecoder(resp.Body).Decode(&res)
				if decodeErr != nil {
					http.Error(w, "manager error decoding response", http.StatusBadRequest)
				} else {
					//update list
					for _, val := range res.Contents {
						cs.SubscriptionData[val.Id] = val
					}

					if subscription, found := cs.SubscriptionData[Client]; found {
						//next
						now := time.Now()
						then := subscription.EndDate
						if then.After(now) {
							next.ServeHTTP(w, r)
						} else {
							http.Error(w, "Customer subscription expired", http.StatusBadRequest)
						}

					} else {
						http.Error(w, "Client has no subscription", http.StatusBadRequest)
					}

				}

			}
		}
	})

}

func FetchSubscriptions(cs *ClientSubscription, c chan string) {
	// if len(cs.SubscriptionData) == 0 {
	headerParams := &HeaderParams{}

	var reqOpt *RequestParams = &RequestParams{
		Service: "manager",
		Path:    "subscriptions",
		Method:  "GET",
		Data:    map[string]interface{}{},
		Header:  headerParams,
	}
	resp, err := CallService(reqOpt)
	if err != nil {
		fmt.Println("manager service down!")
	} else if resp.StatusCode != 200 {
		fmt.Println("manager service not available!")
	} else {
		var res Subscriptions
		decodeErr := json.NewDecoder(resp.Body).Decode(&res)
		if decodeErr != nil {
			fmt.Println("manager error decoding response")
		} else {
			//update list
			for _, val := range res.Contents {
				cs.SubscriptionData[val.Id] = val
			}
		}
	}
	// }
	c <- "ok"
}
