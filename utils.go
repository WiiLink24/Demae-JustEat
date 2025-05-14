package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/WiiLink24/nwc24"
	"github.com/getsentry/sentry-go"
	"github.com/logrusorgru/aurora/v4"
	"log"
	"net/http"
	"strconv"
)

func NewResponse(r *http.Request, w *http.ResponseWriter, xmlType XMLType) *Response {
	wiiNumber, err := strconv.ParseUint(r.Header.Get("X-WiiNo"), 10, 64)
	if err != nil {
		// Failed to parse Wii Number or invalid integer
		(*w).WriteHeader(http.StatusBadRequest)
		return nil
	}

	number := nwc24.LoadWiiNumber(wiiNumber)
	if !number.CheckWiiNumber() {
		// Bad Wii Number
		(*w).WriteHeader(http.StatusBadRequest)
		return nil
	}

	return &Response{
		ResponseFields: demae.KVFieldWChildren{
			XMLName: xml.Name{Local: "response"},
			Value:   nil,
		},
		wiiNumber:           number,
		request:             r,
		writer:              w,
		isMultipleRootNodes: xmlType == 1,
	}
}

func (r *Response) GetHollywoodId() string {
	return strconv.Itoa(int(r.wiiNumber.GetHollywoodID()))
}

// AddCustomType adds a given key by name to a specified structure.
func (r *Response) AddCustomType(customType any) {
	k, ok := r.ResponseFields.(demae.KVFieldWChildren)
	if ok {
		k.Value = append(k.Value, customType)
		r.ResponseFields = k
		return
	}

	// Now check if the fields is an array of any.
	array, ok := r.ResponseFields.([]any)
	if ok {
		r.ResponseFields = append(r.ResponseFields.([]any), array)
	}
}

// AddKVNode adds a given key by name to a specified value, such as <key>value</key>.
func (r *Response) AddKVNode(key string, value string) {
	k, ok := r.ResponseFields.(demae.KVFieldWChildren)
	if !ok {
		return
	}

	k.Value = append(k.Value, demae.KVField{
		XMLName: xml.Name{Local: key},
		Value:   value,
	})

	r.ResponseFields = k
}

// AddKVWChildNode adds a given key by name to a specified value, such as <key><child>...</child></key>.
func (r *Response) AddKVWChildNode(key string, value any) {
	k, ok := r.ResponseFields.(demae.KVFieldWChildren)
	if !ok {
		return
	}

	k.Value = append(k.Value, demae.KVFieldWChildren{
		XMLName: xml.Name{Local: key},
		Value:   []any{value},
	})

	r.ResponseFields = k
}

func (r *Response) toXML() (string, error) {
	var contents string

	if r.isMultipleRootNodes {
		var temp []byte
		var err error
		array, ok := r.ResponseFields.([]any)
		if ok {
			for _, a := range array {
				temp, err = xml.MarshalIndent(a, "", "  ")
				if err != nil {
					return "", err
				}

				contents += string(temp) + "\n"
			}
		} else {
			temp, err = xml.MarshalIndent(r.ResponseFields, "", "  ")
			if err != nil {
				return "", err
			}

			contents += string(temp) + "\n"
		}

		// Now the version and API tags
		version, apiStatus := GenerateVersionAndAPIStatus()
		temp, err = xml.MarshalIndent(version, "", "  ")
		if err != nil {
			return "", err
		}

		contents += string(temp) + "\n"

		temp, err = xml.MarshalIndent(apiStatus, "", "  ")
		if err != nil {
			return "", err
		}

		contents += string(temp)
	} else {
		version, apiStatus := GenerateVersionAndAPIStatus()
		r.AddCustomType(version)
		r.AddCustomType(apiStatus)
		temp, err := xml.MarshalIndent(r.ResponseFields, "", "  ")
		if err != nil {
			return "", err
		}

		contents += string(temp)
	}

	return contents, nil
}

func GenerateVersionAndAPIStatus() (*demae.KVField, *demae.KVFieldWChildren) {
	version := demae.KVField{
		XMLName: xml.Name{Local: "version"},
		Value:   "1",
	}

	apiStatus := demae.KVFieldWChildren{
		XMLName: xml.Name{Local: "apiStatus"},
		Value: []any{
			demae.KVField{
				XMLName: xml.Name{Local: "code"},
				Value:   "0",
			},
		},
	}

	return &version, &apiStatus
}
func PostDiscordWebhook(title, message, url string, color int) {
	theMap := map[string]any{
		"content": nil,
		"embeds": []map[string]any{
			{
				"title":       title,
				"description": message,
				"color":       color,
			},
		},
	}

	jsonData, _ := json.Marshal(theMap)
	_, _ = http.Post(url, "application/json", bytes.NewBuffer(jsonData))
}

// ReportError helps make errors nicer. First it logs the error to Sentry,
// then writes a response for the server to send.
func (r *Response) ReportError(err error) {
	/*if !errors.Is(err, dominos.InvalidCountry) && r.dominos != nil {
		// Write the JSON Dominos sent us to the system.
		_ = os.WriteFile(fmt.Sprintf("errors/%s_%s.json", r.request.URL.Path, r.request.Header.Get("X-WiiNo")), r.dominos.GetResponse(), 0664)
	}*/

	sentry.WithScope(func(s *sentry.Scope) {
		s.SetTag("Wii ID", r.GetHollywoodId())
		sentry.CaptureException(err)
	})

	log.Printf("An error has occurred: %s", aurora.Red(err.Error()))

	errorString := fmt.Sprintf("%s\nWii ID: %s\nWii Number: %s", err.Error(), r.GetHollywoodId(), r.request.Header.Get("X-WiiNo"))
	PostDiscordWebhook("An error has occurred in Demae Domino's!", errorString, config.ErrorWebhook, 16711711)

	// With the new patches I created, we can now send the error to the channel.
	r.AddKVNode("error", err.Error())
}

func printError(w http.ResponseWriter, reason string, code int) {
	http.Error(w, reason, code)
	log.Print("Failed to handle request: ", aurora.Red(reason))
}
