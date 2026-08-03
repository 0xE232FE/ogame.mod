package gameforge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/dop251/goja"
)

const powCaptchaBaseURL = "https://pow-captcha.gameforge.com"

// PowCaptchaProbe is a single browser instrumentation probe from the .pow file.
type PowCaptchaProbe struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Code string `json:"code"`
}

// fetchPowChallenge fetches the .pow container for a pow-captcha challenge.
func fetchPowChallenge(ctx context.Context, client HttpClient, challengeID string) (*PowFile, error) {
	by, err := fetchURL(ctx, client, powCaptchaBaseURL+"/api/challenge/"+challengeID)
	if err != nil {
		return nil, err
	}
	pf, err := ParsePowFile(by)
	if err != nil {
		return nil, err
	}
	return pf, nil
}

// parsePowProbes decodes the double-encoded instrumentation field into probes.
func parsePowProbes(pf *PowFile) ([]PowCaptchaProbe, error) {
	var probes []PowCaptchaProbe
	if err := json.Unmarshal([]byte(pf.Instrumentation), &probes); err != nil {
		return nil, err
	}
	return probes, nil
}

// evaluatePowProbe runs a probe's JS in a simulated browser environment and
// returns its result. bitwise/prototype probes are evaluated exactly; canvas/dom
// probes return deterministic plausible values (a headless client cannot render
// fonts like a real browser).
func evaluatePowProbe(code string) (int64, error) {
	rt := goja.New()
	_ = rt.Set("navigator", map[string]interface{}{"userAgent": "Mozilla/5.0", "language": "en-US"})
	_ = rt.Set("window", map[string]interface{}{})
	_ = rt.Set("document", map[string]interface{}{})
	_ = rt.Set("setTimeout", func(interface{}, int64) {})
	_ = rt.Set("requestAnimationFrame", func(interface{}) {})
	evalFn := func(_ goja.FunctionCall) goja.Value { return goja.Undefined() }
	_ = rt.Set("eval", rt.ToValue(evalFn))
	_, _ = rt.RunString(`Function.prototype.toString = function(){ return "[native code]"; };`)
	_, _ = rt.RunString(`
		document.createElement = function(tag){
			if(tag === 'canvas'){
				return {
					width: 0, height: 0,
					getContext: function(){ return {
						font: '', fillStyle: '',
						_fillText: '',
						fillText: function(t){ this._fillText = t; },
						getImageData: function(x,y,w,h){
							var d = new Uint8Array(w*h*4);
							var t = this._fillText || '';
							var seed = 0;
							for (var i=0;i<t.length;i++){ seed = ((seed<<5)-seed+t.charCodeAt(i))|0; }
							for (var i=0;i<w*h*4;i+=4){ d[i] = (seed + i) & 0xff; d[i+1] = (seed >> 8) & 0xff; d[i+2] = 0; d[i+3] = 255; }
							return { data: d };
						}
					};}
				};
			}
			if(tag === 'div'){ return { style: {}, offsetHeight: 125, offsetWidth: 250 }; }
			return {};
		};
		document.body = { appendChild: function(){}, removeChild: function(){} };
	`)
	val, err := rt.RunString("(function(){ " + code + " })()")
	if err != nil {
		return 0, err
	}
	return val.ToInteger(), nil
}

// SolvePowCaptcha solves a gf-pow-captcha challenge: fetches the .pow container,
// solves the sha-256 challenges, evaluates the instrumentation probes and submits
// both to mark the challenge as solved.
func SolvePowCaptcha(ctx context.Context, client HttpClient, challengeID string) error {
	pf, err := fetchPowChallenge(ctx, client, challengeID)
	if err != nil {
		return errors.New("failed to fetch pow challenge: " + err.Error())
	}

	pow := make([]map[string]string, len(pf.Pow.Challenges))
	for i, c := range pf.Pow.Challenges {
		nonce := SolvePow(c.Salt, c.Target)
		pow[i] = map[string]string{"salt": c.Salt, "nonce": strconv.FormatInt(nonce, 10)}
	}

	probes, err := parsePowProbes(pf)
	if err != nil {
		return errors.New("failed to parse pow instrumentation: " + err.Error())
	}
	instrumentation := make([]int64, len(probes))
	for i, p := range probes {
		if v, err := evaluatePowProbe(p.Code); err == nil {
			instrumentation[i] = v
		}
	}

	payload := map[string]interface{}{
		"pow":             pow,
		"instrumentation": instrumentation,
		"metrics":         map[string]interface{}{"solver": map[string]interface{}{}},
	}
	pj, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, powCaptchaBaseURL+"/api/challenge/"+challengeID, bytes.NewReader(pj))
	if err != nil {
		return err
	}
	req.Header.Set(contentTypeHeaderKey, applicationJson)
	req.Header.Set(acceptEncodingHeaderKey, gzipEncoding)
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	by, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result struct {
		Status string `json:"status"`
	}
	_ = json.Unmarshal(by, &result)
	if resp.StatusCode != http.StatusOK || result.Status != "solved" {
		return fmt.Errorf("pow challenge verification failed (status %d): %s", resp.StatusCode, string(by))
	}
	return nil
}

// isPowCaptchaChallenge reports whether the challenge served by gameforge is a
// proof-of-work captcha (gf-pow-captcha) rather than the image-drop captcha.
func isPowCaptchaChallenge(ctx context.Context, client HttpClient, challengeID string) (bool, error) {
	by, err := fetchURL(ctx, client, getChallengeURL(challengeBaseURL, challengeID))
	if err != nil {
		return false, err
	}
	var info struct {
		Type string `json:"type"`
	}
	_ = json.Unmarshal(by, &info)
	return info.Type == "gf-pow-captcha", nil
}
