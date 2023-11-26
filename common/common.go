package common

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type RequestParams struct {
	Service string
	Path    string
	Method  string
	Data    map[string]interface{}
	Header  *HeaderParams
}

type User struct {
	Id                string `json:"_id"`
	Birthday          time.Time
	Capabilities      []string
	Email             string
	EmailConfirmation bool
	FirstName         string
	LastName          string
	Gender            string
	Meta              []map[string]interface{}
	Picture           string
	Roles             []string
}

type Credentials struct {
	Authentication bool
	Authorization  bool
	User           User
	Err            error
}

type HeaderParams struct {
	Client         string
	Service        string
	ServiceToken   string
	UserId         string
	AccessToken    string
	AcceptLanguage string
	// Expire         string
	Expire                 time.Time `bson:"expire, omitempty" `
	VideoConvertingOptions string
}

type Option struct {
	Id    string
	Name  string
	Value interface{}
}

type OptionResult struct {
	Option Option
}

func CallService(reqOpt *RequestParams) (*http.Response, error) {
	serviceUrl := GetServiceUrl(reqOpt.Service)
	path := reqOpt.Path
	url := fmt.Sprintf("%s/%s", serviceUrl, path)
	// fmt.Println("CallService:", url, ", Method:", reqOpt.Method)
	j, _ := json.Marshal(reqOpt.Data)
	req, err := http.NewRequest(reqOpt.Method, url, bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}
	req.Header.Add("x-client", reqOpt.Header.Client)
	req.Header.Add("x-service", reqOpt.Header.Service)
	req.Header.Add("x-service-token", reqOpt.Header.ServiceToken)
	req.Header.Add("x-user-id", reqOpt.Header.UserId)
	req.Header.Add("x-access-token", reqOpt.Header.AccessToken)
	req.Header.Add("accept-language", reqOpt.Header.AcceptLanguage)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, errResp := client.Do(req)
	if errResp != nil {
		return nil, errResp
	}

	// fmt.Println("status", resp.StatusCode)

	return resp, nil

}

func ExtractHeaderParams(r *http.Request) *HeaderParams {
	p := &HeaderParams{}
	p.Client = r.Header.Get("x-client")
	p.Service = r.Header.Get("x-service")
	p.ServiceToken = r.Header.Get("x-service-token")
	p.UserId = r.Header.Get("x-user-id")
	p.AccessToken = r.Header.Get("x-access-token")
	p.AcceptLanguage = r.Header.Get("accept-language")
	p.VideoConvertingOptions = r.Header.Get("x-video-converting-options")

	if r.Header.Get("x-expire") != "" {
		expire, err := time.Parse(time.RFC3339, r.Header.Get("x-expire"))
		if err != nil {
			fmt.Println(err)
		}
		p.Expire = expire
	}

	return p

}

func GetServiceUrl(name string) string {

	var services = make(map[string]string)

	// manager
	services["manager"] = getURL("MANAGER_URL", "3030")

	// options
	services["options"] = getURL("OPTIONS_URL", "3200")

	// users
	services["users"] = getURL("USERS_URL", "3201")

	//email
	services["email"] = getURL("EMAIL_URL", "3202")

	// elearning
	services["elearning"] = getURL("ELEARNING_URL", "3203")

	// files
	services["files"] = getURL("FILES_URL", "3204")

	//serviceAuth
	services["serviceAuth"] = getURL("SERVICE_AUTH_URL", "3209")

	//notifications
	services["notifications"] = getURL("NOTIFICATIONS_URL", "3205")

	//contact
	services["contact"] = getURL("CONTACT_URL", "3206")

	//cms
	services["cms"] = getURL("CMS_URL", "3207")

	//ticketing
	services["ticketing"] = getURL("TICKETING_URL", "3210")

	//types
	services["types"] = getURL("TYPES_URL", "3212")

	//survey
	services["survey"] = getURL("SURVEY_URL", "3213")

	//paymentSmartDubaiHelper
	services["paymentSmartDubaiHelper"] = getURL("SMART_DUBAI_HELPER_PROXY_URL", "9735")

	//payment
	services["payment"] = getURL("PAYMENT_URL", "3214")

	//newsletter
	services["newsletter"] = getURL("NEWSLETTER_URL", "3215")

	//imageBuilder
	services["imageBuilder"] = getURL("IMAGE_BUILDER_URL", "3216")

	//tahkeem
	services["tahkeem"] = getURL("TAHKEEM_URL", "3217")

	//audit
	services["audit"] = getURL("AUDIT_URL", "3218")

	//backup
	services["backup"] = getURL("BACKUP_URL", "3223")

	return services[name]

}

func getURL(ename string, port string) string {
	if url := os.Getenv(ename); url != "" {
		return url
	} else {
		return "http://localhost:" + port
	}
}

func Guard(c chan Credentials, h *HeaderParams, restrictions []string) {
	// fmt.Println("start Guard: ", time.Now())
	var reqOpt *RequestParams = &RequestParams{}
	var resp *http.Response
	var err error
	var res Credentials
	if h.ServiceToken != "" {
		reqOpt.Service = "serviceAuth"
		reqOpt.Path = "verify"
		reqOpt.Method = "POST"
		reqOpt.Data = map[string]interface{}{
			"token": h.ServiceToken,
		}

		reqOpt.Header = h

		resp, err = CallService(reqOpt)
		if err != nil || resp.StatusCode != 200 {
			c <- Credentials{Err: errors.New("not authenticated")}
			return
		}

		/*
		 /*	todo maybe check for decode error
		*/

		json.NewDecoder(resp.Body).Decode(&res)

		c <- res

		// TODO see one more case //

	} else {
		reqOpt.Service = "users"
		reqOpt.Path = "guard"
		reqOpt.Method = "POST"
		reqOpt.Data = map[string]interface{}{"restrictions": restrictions}
		reqOpt.Header = h
		resp, err = CallService(reqOpt)
		// fmt.Println("send users/guard Done: ", time.Now())
		if err != nil || resp.StatusCode != 200 {
			c <- Credentials{Err: errors.New("not authenticated")}
			return
		}

		json.NewDecoder(resp.Body).Decode(&res)

		c <- res

	}

}

type CommonContextKey string

// MiddlewareHandler Type
type Middleware func(http.Handler) http.Handler

// MiddlewaresHandler takes Handler funcs and chains them to the main handler
func MiddlewaresHandler(handler http.Handler, middlewares ...Middleware) http.Handler {
	// The loop is reversed so the middlewares gets executed in the same
	// order as provided in the array.
	for i := len(middlewares); i > 0; i-- {
		handler = middlewares[i-1](handler)
	}
	return handler
}

func GuardMiddleware(restructions []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := make(chan Credentials)
			// fmt.Println("start guardMiddleware Done: ", time.Now())
			headerParams := ExtractHeaderParams(r)
			go Guard(c, headerParams, restructions)
			credentials := <-c
			if credentials.Err != nil {
				http.Error(w, credentials.Err.Error(), http.StatusBadRequest)
				return
			}
			ctx := context.WithValue(r.Context(), CommonContextKey("user"), credentials.User)
			// fmt.Println("next of  guardMiddleware: ", time.Now())
			next.ServeHTTP(w, r.Clone(ctx))
		})
	}
}

// func Log() Middleware {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

func GetOptionValue(option string, h *HeaderParams) (interface{}, error) {
	var reqOpt *RequestParams = &RequestParams{
		Service: "options",
		Path:    option,
		Method:  "GET",
		Data:    map[string]interface{}{},
		Header:  h,
	}
	resp, err := CallService(reqOpt)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 200 {
		var res OptionResult
		decodeErr := json.NewDecoder(resp.Body).Decode(&res)
		if decodeErr != nil {
			return nil, decodeErr
		}
		return res.Option.Value, nil
	}

	return nil, nil
}

func GetServiceToken(service string, h *HeaderParams) (string, error) {
	var reqOpt *RequestParams = &RequestParams{
		Service: "serviceAuth",
		Path:    "issue/" + service,
		Method:  "POST",
		Data:    map[string]interface{}{},
		Header:  h,
	}
	resp, err := CallService(reqOpt)
	if err != nil {
		return "", err
	}

	type ServiceToken struct {
		Token string
	}

	if resp.StatusCode == 200 {
		var res ServiceToken
		decodeErr := json.NewDecoder(resp.Body).Decode(&res)
		if decodeErr != nil {
			return "", decodeErr
		}
		return res.Token, nil
	}

	return "", nil
}

func Download(url string, dest string, h *HeaderParams) error {

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	if h != nil {
		req.Header.Add("x-client", h.Client)
		req.Header.Add("x-service", h.Service)
		req.Header.Add("x-service-token", h.ServiceToken)
		req.Header.Add("x-user-id", h.UserId)
		req.Header.Add("x-access-token", h.AccessToken)
		req.Header.Add("accept-language", h.AcceptLanguage)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func Upload(url string, path string, h *HeaderParams) (*http.Response, error) {

	pr, pw := io.Pipe()
	mpw := multipart.NewWriter(pw)

	errchan := make(chan error)

	go func() {

		defer close(errchan)
		defer mpw.Close()
		defer pw.Close()

		w, err := mpw.CreateFormFile("file", filepath.Base(path))
		if err != nil {
			errchan <- err
			return
		}

		in, err := os.Open(path)
		if err != nil {
			errchan <- err
			return
		}
		defer in.Close()

		if written, err := io.Copy(w, in); err != nil {
			errchan <- fmt.Errorf("error copying %s (%d bytes written): %v", path, written, err)
			return
		}

		if err := mpw.Close(); err != nil {
			errchan <- err
			return
		}

	}()

	req, err := http.NewRequest("POST", url, pr)
	if err != nil {
		return nil, err
	}
	req.Header.Add("x-client", h.Client)
	req.Header.Add("x-service", h.Service)
	req.Header.Add("x-service-token", h.ServiceToken)
	req.Header.Add("x-user-id", h.UserId)
	req.Header.Add("x-access-token", h.AccessToken)
	req.Header.Add("accept-language", h.AcceptLanguage)
	req.Header.Set("Content-Type", mpw.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)

	merr := <-errchan

	if err != nil || merr != nil {
		fmt.Println("http error:, multipart error: ", err, merr)
	}

	if err != nil {
		return nil, err
	}
	if merr != nil {
		return nil, merr
	}

	return resp, nil
}
