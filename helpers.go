package firebase_tools

import (
	"log"
	"net/http"
)

// CloseRespBody closes the body of the supplied HTTP response
func CloseRespBody(resp *http.Response) {
	if resp != nil {
		err := resp.Body.Close()
		if err != nil {
			log.Println("unable to close response body for request made to ", resp.Request.RequestURI)
		}
	}
}
