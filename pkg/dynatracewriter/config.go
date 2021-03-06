package dynatracewriter

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"
    "fmt"
	"github.com/kubernetes/helm/pkg/strvals"
	"go.k6.io/k6/lib/types"
    "gopkg.in/guregu/null.v3"
)

const (
	defaultDynatraceTimeout = time.Minute
	defaultFlushPeriod       = time.Second
	defaultMetricPrefix      = "k6."
	defaultDynatraceMetricEndPoint ="/api/v2/metrics/ingest"
)

type Config struct {
	Url string `json:"url" envconfig:"K6_DYNATRACE_URL"` // here, in the name of env variable, we assume that we won't need to distinguish between remote write URL vs remote read URL
    Headers map[string]string `json:"headers" envconfig:"K6_DYNATRACE_HEADER"`
	InsecureSkipTLSVerify null.Bool   `json:"insecureSkipTLSVerify" envconfig:"K6_DYNATRACE_INSECURE_SKIP_TLS_VERIFY"`
	CACert                null.String `json:"caCertFile" envconfig:"K6_CA_CERT_FILE"`
	ApiToken     null.String `json:"apitoken" envconfig:"K6_DYNATRACE_APITOKEN"`
	FlushPeriod types.NullDuration `json:"flushPeriod" envconfig:"K6_DYNATRACE_FLUSH_PERIOD"`
	KeepTags    null.Bool `json:"keepTags" envconfig:"K6_KEEP_TAGS"`
	KeepNameTag null.Bool `json:"keepNameTag" envconfig:"K6_KEEP_NAME_TAG"`
	KeepUrlTag  null.Bool `json:"keepUrlTag" envconfig:"K6_KEEP_URL_TAG"`
}

func NewConfig() Config {
	return Config{
		Url:                   "https://dynatrace.live.com",
		InsecureSkipTLSVerify: null.BoolFrom(true),
		CACert:                null.NewString("", false),
        ApiToken:              null.NewString("", false),
		FlushPeriod:           types.NullDurationFrom(defaultFlushPeriod),
		KeepTags:              null.BoolFrom(true),
		KeepNameTag:           null.BoolFrom(false),
		KeepUrlTag:            null.BoolFrom(true),
		Headers:               make(map[string]string),
	}
}

func (conf Config) ConstructConfig() (*Config, error) {
	// TODO: consider if the auth logic should be enforced here
	// (e.g. if insecureSkipTLSVerify is switched off, then check for non-empty certificate file and auth, etc.)

	u, err := url.Parse(conf.Url+defaultDynatraceMetricEndPoint)
	if err != nil {
		return nil, err
	}
    if len(conf.ApiToken.String) == 0 {
       return nil, fmt.Errorf("The Dynatrace API token can not been empty or Null")
    } else {
        conf.Headers["Content-Type"] = "text/plain; charset=utf-8"
        conf.Headers["Authorization"] ="Api-Token " + conf.ApiToken.String
        conf.Headers["accept"] = "*/*"
    }
     conf.Url= u.String()

	return &conf, nil
}

// From here till the end of the file partial duplicates waiting for config refactor (k6 #883)

func (base Config) Apply(applied Config) Config {


	if len(applied.Url)>0 {
		base.Url = applied.Url
	}

	if applied.InsecureSkipTLSVerify.Valid {
		base.InsecureSkipTLSVerify = applied.InsecureSkipTLSVerify
	}

	if applied.CACert.Valid {
		base.CACert = applied.CACert
	}

	if applied.ApiToken.Valid {
		base.ApiToken = applied.ApiToken
	}



	if applied.FlushPeriod.Valid {
		base.FlushPeriod = applied.FlushPeriod
	}

	if applied.KeepTags.Valid {
		base.KeepTags = applied.KeepTags
	}

	if applied.KeepNameTag.Valid {
		base.KeepNameTag = applied.KeepNameTag
	}

	if applied.KeepUrlTag.Valid {
		base.KeepUrlTag = applied.KeepUrlTag
	}

	if len(applied.Headers) > 0 {
		for k, v := range applied.Headers {
			base.Headers[k] = v
		}
	}

	return base
}

// ParseArg takes an arg string and converts it to a config
func ParseArg(arg string) (Config, error) {
	var c Config
	params, err := strvals.Parse(arg)
	if err != nil {
		return c, err
	}

	if v, ok := params["url"].(string); ok {
		c.Url = v
	}

	if v, ok := params["insecureSkipTLSVerify"].(bool); ok {
		c.InsecureSkipTLSVerify = null.BoolFrom(v)
	}

	if v, ok := params["caCertFile"].(string); ok {
		c.CACert = null.StringFrom(v)
	}

	if v, ok := params["apitoken"].(string); ok {
		c.ApiToken = null.StringFrom(v)
	}


	if v, ok := params["flushPeriod"].(string); ok {
		if err := c.FlushPeriod.UnmarshalText([]byte(v)); err != nil {
			return c, err
		}
	}

	if v, ok := params["keepTags"].(bool); ok {
		c.KeepTags = null.BoolFrom(v)
	}

	if v, ok := params["keepNameTag"].(bool); ok {
		c.KeepNameTag = null.BoolFrom(v)
	}

	if v, ok := params["keepUrlTag"].(bool); ok {
		c.KeepUrlTag = null.BoolFrom(v)
	}

	c.Headers = make(map[string]string)
	if v, ok := params["headers"].(map[string]interface{}); ok {
		for k, v := range v {
			if v, ok := v.(string); ok {
				c.Headers[k] = v
			}
		}
	}

	return c, nil
}

// GetConsolidatedConfig combines {default config values + JSON config +
// environment vars + arg config values}, and returns the final result.
func GetConsolidatedConfig(jsonRawConf json.RawMessage, env map[string]string, arg string) (Config, error) {
	result := NewConfig()
	if jsonRawConf != nil {
		jsonConf := Config{}
		if err := json.Unmarshal(jsonRawConf, &jsonConf); err != nil {
			return result, err
		}
		result = result.Apply(jsonConf)
	}

	getEnvBool := func(env map[string]string, name string) (null.Bool, error) {
		if v, vDefined := env[name]; vDefined {
			if b, err := strconv.ParseBool(v); err != nil {
				return null.NewBool(false, false), err
			} else {
				return null.BoolFrom(b), nil
			}
		}
		return null.NewBool(false, false), nil
	}

	getEnvMap := func(env map[string]string, prefix string) map[string]string {
		result := make(map[string]string)
		for ek, ev := range env {
			if strings.HasPrefix(ek, prefix) {
				k := strings.TrimPrefix(ek, prefix)
				result[k] = ev
			}
		}
		return result
	}

	// envconfig is not processing some undefined vars (at least duration) so apply them manually
	if flushPeriod, flushPeriodDefined := env["K6_DYNATRACE_FLUSH_PERIOD"]; flushPeriodDefined {
		if err := result.FlushPeriod.UnmarshalText([]byte(flushPeriod)); err != nil {
			return result, err
		}
	}



	if url, urlDefined := env["K6_DYNATRACE_URL"]; urlDefined {
		result.Url =url
	}

	if b, err := getEnvBool(env, "K6_DYNATRACE_INSECURE_SKIP_TLS_VERIFY"); err != nil {
		return result, err
	} else {
		if b.Valid {
			// apply only if valid, to keep default option otherwise
			result.InsecureSkipTLSVerify = b
		}
	}

	if ca, caDefined := env["K6_CA_CERT_FILE"]; caDefined {
		result.CACert = null.StringFrom(ca)
	}

	if apitoken, userDefined := env["K6_DYNATRACE_APITOKEN"]; userDefined {
		result.ApiToken = null.StringFrom(apitoken)
	}


	if b, err := getEnvBool(env, "K6_KEEP_TAGS"); err != nil {
		return result, err
	} else {
		if b.Valid {
			result.KeepTags = b
		}
	}

	if b, err := getEnvBool(env, "K6_KEEP_NAME_TAG"); err != nil {
		return result, err
	} else {
		if b.Valid {
			result.KeepNameTag = b
		}
	}

	if b, err := getEnvBool(env, "K6_KEEP_URL_TAG"); err != nil {
		return result, err
	} else {
		if b.Valid {
			result.KeepUrlTag = b
		}
	}

	envHeaders := getEnvMap(env, "K6_DYNATRACE_HEADER")
	for k, v := range envHeaders {
		result.Headers[k] = v
	}

	if arg != "" {
		argConf, err := ParseArg(arg)
		if err != nil {
			return result, err
		}

		result = result.Apply(argConf)
	}

	return result, nil
}