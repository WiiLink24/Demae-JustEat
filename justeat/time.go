package justeat

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	_ "time/tzdata"

	"github.com/WiiLink24/DemaeJustEat/demae"
	"github.com/WiiLink24/DemaeJustEat/logger"
)

func (j *JEClient) getLocalizedTimeLocation() (*time.Location, error) {
	return time.LoadLocation(timeZones[j.Country])
}

func (j *JEClient) getLocalizedTime(value string) (time.Time, error) {
	zone, err := j.getLocalizedTimeLocation()
	if err != nil {
		return time.Time{}, err
	}

	_t, err := time.Parse("2006-01-02T15:04:05Z", value)
	if err != nil {
		return time.Time{}, err
	}

	return _t.In(zone), nil
}

func (j *JEClient) getAvailableTimes(basketId string) (map[string]any, error) {
	resp, err := j.httpGet(fmt.Sprintf("%s/checkout/%s/%s/fulfilment/availabletimes", j.KongAPIURL, strings.ToLower(string(j.Country)), basketId))
	if err != nil {
		return nil, err
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error(_Menu, err.Error())
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)

	var availability map[string]any
	err = json.Unmarshal(body, &availability)
	return availability, err
}

func (j *JEClient) GetAvailableTimes(basketId string) ([]demae.KVFieldWChildren, error) {
	availability, err := j.getAvailableTimes(basketId)
	if err != nil {
		return nil, err
	}

	var times []demae.KVFieldWChildren
	for i, _time := range availability["times"].([]any) {
		_t, err := j.getLocalizedTime(_time.(map[string]any)["from"].(string))
		if err != nil {
			return nil, err
		}

		times = append(times, demae.KVFieldWChildren{
			XMLName: xml.Name{Local: "option"},
			Value: []any{
				demae.KVField{
					XMLName: xml.Name{Local: "id"},
					Value:   i,
				},
				demae.KVField{
					XMLName: xml.Name{Local: "name"},
					Value:   _t.Format("2006-01-02 15:04:05"),
				},
			},
		})
	}

	return times, err
}
