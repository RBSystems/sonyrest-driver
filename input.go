package sonyrest

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/byuoitav/common/log"
	"github.com/byuoitav/common/nerr"
)

// GetInput gets the input that is currently being shown on the TV
func (t *TV) GetAudioVideoInputs(ctx context.Context) (map[string]string, error) {
	output := make(map[string]string)

	pwrState, err := t.GetPower(ctx)
	if err != nil {
		return output, err
	}
	if !pwrState {
		return output, nil
	}

	payload := SonyTVRequest{
		Params:  []map[string]interface{}{},
		Method:  "getPlayingContentInfo",
		ID:      1,
		Version: "1.0",
	}

	response, err := t.PostHTTPWithContext(ctx, "avContent", payload)
	if err != nil {
		return output, err
	}

	var outputStruct SonyAVContentResponse
	err = json.Unmarshal(response, &outputStruct)
	if err != nil || len(outputStruct.Result) < 1 {
		return output, err
	}
	//we need to parse the response for the value

	log.L.Debugf("%+v", outputStruct)

	regexStr := `extInput:(.*?)\?port=(.*)`
	re := regexp.MustCompile(regexStr)

	matches := re.FindStringSubmatch(outputStruct.Result[0].URI)
	output[""] = fmt.Sprintf("%v!%v", matches[1], matches[2])

	log.L.Infof("Current Input for %s: %s", t.Address, output[""])

	return output, nil
}

func (t *TV) SetAudioVideoInput(ctx context.Context, output, input string) error {
	log.L.Infof("Switching input for %s to %s ...", t.Address, input)

	splitPort := strings.Split(input, "!")

	params := make(map[string]interface{})
	params["uri"] = fmt.Sprintf("extInput:%s?port=%s", splitPort[0], splitPort[1])

	err := t.BuildAndSendPayload(ctx, t.Address, "avContent", "setPlayContent", params)
	if err != nil {
		return err
	}

	log.L.Debugf("Done.")
	return nil
}

// GetActiveSignal determines if the current input on the TV is active or not
func (t *TV) GetActiveSignal(ctx context.Context, port string) (bool, error) {
	var activeSignal bool

	payload := SonyTVRequest{
		Params:  []map[string]interface{}{},
		Method:  "getCurrentExternalInputsStatus",
		ID:      1,
		Version: "1.1",
	}

	response, err := t.PostHTTPWithContext(ctx, "avContent", payload)
	if err != nil {
		return activeSignal, nerr.Translate(err)
	}

	var outputStruct SonyMultiAVContentResponse
	err = json.Unmarshal(response, &outputStruct)
	if err != nil || len(outputStruct.Result) < 1 {
		return activeSignal, nerr.Translate(err)
	}
	//we need to parse the response for the value

	log.L.Debugf("%+v", outputStruct)

	regexStr := `extInput:(.*?)\?port=(.*)`
	re := regexp.MustCompile(regexStr)

	for _, result := range outputStruct.Result[0] {
		if result.Status == "true" {
			matches := re.FindStringSubmatch(result.URI)
			tempActive := fmt.Sprintf("%v!%v", matches[1], matches[2])

			activeSignal = (tempActive == port)
		}
	}

	return activeSignal, nil
}
