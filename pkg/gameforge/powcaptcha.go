package gameforge

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/alaingilbert/ogame/pkg/device"
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
// probes return values seeded from the device fingerprint so that they stay
// consistent with the blackbox sent during the login.
func evaluatePowProbe(code string, fpSeed int64) (int64, error) {
	rt := goja.New()
	_ = rt.Set("navigator", map[string]interface{}{"userAgent": "Mozilla/5.0", "language": "en-US"})
	_ = rt.Set("window", map[string]interface{}{})
	_ = rt.Set("document", map[string]interface{}{})
	_ = rt.Set("setTimeout", func(interface{}, int64) {})
	_ = rt.Set("requestAnimationFrame", func(interface{}) {})
	evalFn := func(_ goja.FunctionCall) goja.Value { return goja.Undefined() }
	_ = rt.Set("eval", rt.ToValue(evalFn))
	_, _ = rt.RunString(`Function.prototype.toString = function(){ return "[native code]"; };`)
	_ = rt.Set("_fpSeed", fpSeed)
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
							var seed = _fpSeed;
							for (var i=0;i<t.length;i++){ seed = ((seed<<5)-seed+t.charCodeAt(i))|0; }
							for (var i=0;i<w*h*4;i+=4){ d[i] = (seed + i) & 0xff; d[i+1] = (seed >> 8) & 0xff; d[i+2] = (seed >> 16) & 0xff; d[i+3] = 255; }
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

// fingerprintSeed derives a stable seed from the device fingerprint so that the
// pow-captcha instrumentation is consistent with the blackbox sent at login.
func fingerprintSeed(fp *device.JsFingerprint) int64 {
	if fp == nil {
		return 0
	}
	h := sha256.New()
	_ = binary.Write(h, binary.LittleEndian, int64(fp.Canvas2DInfo))
	_ = binary.Write(h, binary.LittleEndian, int64(fp.ScreenWidth))
	_ = binary.Write(h, binary.LittleEndian, int64(fp.ScreenHeight))
	h.Write([]byte(fp.FontsHash))
	h.Write([]byte(fp.WebglInfo))
	h.Write([]byte(fp.AudioCtxHash))
	sum := h.Sum(nil)
	return int64(binary.LittleEndian.Uint64(sum[:8]))
}

// SolvePowCaptcha solves a gf-pow-captcha challenge: fetches the .pow container,
// solves the sha-256 challenges, evaluates the instrumentation probes (seeded from
// the device fingerprint) and submits both to mark the challenge as solved.
func SolvePowCaptcha(ctx context.Context, device Device, challengeID string) error {
	pf, err := fetchPowChallenge(ctx, device, challengeID)
	if err != nil {
		return errors.New("failed to fetch pow challenge: " + err.Error())
	}

	pow := make([]map[string]string, len(pf.Pow.Challenges))
	for i, c := range pf.Pow.Challenges {
		nonce := SolvePow(c.Salt, c.Target)
		pow[i] = map[string]string{"salt": c.Salt, "nonce": strconv.FormatInt(nonce, 10)}
	}

	var fpSeed int64
	if bb, err := device.GetBlackbox(); err == nil {
		if fp, ferr := parseFingerprint(bb); ferr == nil {
			fpSeed = fingerprintSeed(fp)
		}
	}

	probes, err := parsePowProbes(pf)
	if err != nil {
		return errors.New("failed to parse pow instrumentation: " + err.Error())
	}
	instrumentation := make([]int64, len(probes))
	for i, p := range probes {
		if v, err := evaluatePowProbe(p.Code, fpSeed); err == nil {
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

	resp, err := device.Do(req)
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

// parseFingerprint decodes a "tra:"-prefixed blackbox into a JsFingerprint.
func parseFingerprint(blackbox string) (*device.JsFingerprint, error) {
	enc := strings.TrimPrefix(blackbox, blackboxPrefix)
	dec, err := device.DecryptBlackbox(enc)
	if err != nil {
		return nil, err
	}
	return device.ParseBlackbox(dec)
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
