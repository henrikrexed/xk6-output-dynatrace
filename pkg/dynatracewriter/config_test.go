package dynatracewriter

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.k6.io/k6/lib/types"
	"gopkg.in/guregu/null.v3"
)

func TestApply(t *testing.T) {
	t.Parallel()

	fullConfig := Config{
		Url:                   "some-url",
		InsecureSkipTLSVerify: null.BoolFrom(false),
		CACert:                null.StringFrom("some-file"),
		ApiToken:              null.StringFrom("user"),
		FlushPeriod:           types.NullDurationFrom(10 * time.Second),
		Headers: map[string]string{
			"X-Header": "value",
		},
	}

	// Defaults should be overwritten by valid values
	c := NewConfig()
	c = c.Apply(fullConfig)
	assert.Equal(t, fullConfig.Url, c.Url)
	assert.Equal(t, fullConfig.InsecureSkipTLSVerify, c.InsecureSkipTLSVerify)
	assert.Equal(t, fullConfig.CACert, c.CACert)
	assert.Equal(t, fullConfig.ApiToken, c.ApiToken)
	assert.Equal(t, fullConfig.FlushPeriod, c.FlushPeriod)
	assert.Equal(t, fullConfig.Headers, c.Headers)

	// Defaults shouldn't be impacted by invalid values
	c = NewConfig()
	c = c.Apply(Config{
		ApiToken:              null.NewString("user", false),
		InsecureSkipTLSVerify: null.NewBool(false, false),
	})
	assert.Equal(t, false, c.ApiToken.Valid)
	assert.Equal(t, true, c.InsecureSkipTLSVerify.Valid)
}

func TestConfigParseArg(t *testing.T) {
	t.Parallel()

	c, err := ParseArg("url=https://bix24852.dev.dynatracelabs.com")
	assert.Nil(t, err)
	assert.Equal(t, "https://bix24852.dev.dynatracelabs.com", c.Url)

	c, err = ParseArg("url=https://bix24852.dev.dynatracelabs.com,insecureSkipTLSVerify=false")
	assert.Nil(t, err)
	assert.Equal(t, "https://bix24852.dev.dynatracelabs.com", c.Url)
	assert.Equal(t, null.BoolFrom(false), c.InsecureSkipTLSVerify)

	c, err = ParseArg("url=https://bix24852.dev.dynatracelabs.com,caCertFile=f.crt")
	assert.Nil(t, err)
	assert.Equal(t, "https://bix24852.dev.dynatracelabs.com", c.Url)
	assert.Equal(t, null.StringFrom("f.crt"), c.CACert)

	c, err = ParseArg("url=https://bix24852.dev.dynatracelabs.com,insecureSkipTLSVerify=false,caCertFile=f.crt,apitoken=mytoken")
	assert.Nil(t, err)
	assert.Equal(t, "https://bix24852.dev.dynatracelabs.com", c.Url)
	assert.Equal(t, null.BoolFrom(false), c.InsecureSkipTLSVerify)
	assert.Equal(t, null.StringFrom("f.crt"), c.CACert)
	assert.Equal(t, null.StringFrom("mytoken"), c.ApiToken)

	c, err = ParseArg("url=https://bix24852.dev.dynatracelabs.com,flushPeriod=2s")
	assert.Nil(t, err)
	assert.Equal(t, "https://bix24852.dev.dynatracelabs.com", c.Url)
	assert.Equal(t, types.NullDurationFrom(time.Second*2), c.FlushPeriod)

	c, err = ParseArg("url=https://bix24852.dev.dynatracelabs.com,headers.X-Header=value")
	assert.Nil(t, err)
	assert.Equal(t, "https://bix24852.dev.dynatracelabs.com", c.Url)
	assert.Equal(t, map[string]string{"X-Header": "value"}, c.Headers)
}

func TestGetConsolidatedConfig(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		jsonRaw   json.RawMessage
		env       map[string]string
		arg       string
		config    Config
		errString string
	}{
		"json_success": {
			jsonRaw: json.RawMessage(fmt.Sprintf("{\"url\":\"%s\"}", "https://bix24852.dev.dynatracelabs.com")),
			env:     nil,
			arg:     "",
			config: Config{
				Url:                   "https://bix24852.dev.dynatracelabs.com",
				InsecureSkipTLSVerify: null.BoolFrom(true),
				CACert:                null.NewString("", false),
				ApiToken:              null.NewString("", false),
				FlushPeriod:           types.NullDurationFrom(defaultFlushPeriod),
				KeepTags:              null.BoolFrom(true),
				KeepNameTag:           null.BoolFrom(false),
				KeepUrlTag:            null.BoolFrom(true),
				Headers:               make(map[string]string),
			},
			errString: "",
		},
		"mixed_success": {
			jsonRaw: json.RawMessage(fmt.Sprintf("{\"url\":\"%s\"}", "https://bix24852.dev.dynatracelabs.com")),
			env:     map[string]string{"K6_DYNATRACE_INSECURE_SKIP_TLS_VERIFY": "false", "K6_DYNATRACE_APITOKEN": "u"},
			arg:     "apitoken=user",
			config: Config{
				Url:                   "https://bix24852.dev.dynatracelabs.com",
				InsecureSkipTLSVerify: null.BoolFrom(false),
				CACert:                null.NewString("", false),
				ApiToken:              null.StringFrom("user"),
				FlushPeriod:           types.NullDurationFrom(defaultFlushPeriod),
				KeepTags:              null.BoolFrom(true),
				KeepNameTag:           null.BoolFrom(false),
				KeepUrlTag:            null.BoolFrom(true),
				Headers:               make(map[string]string),
			},
			errString: "",
		},
		"invalid_duration": {
			jsonRaw:   json.RawMessage(fmt.Sprintf("{\"url\":\"%s\"}", "https://bix24852.dev.dynatracelabs.com")),
			env:       map[string]string{"K6_DYNATRACE_FLUSH_PERIOD": "d"},
			arg:       "",
			config:    Config{},
			errString: "strconv.ParseInt",
		},
		"invalid_insecureSkipTLSVerify": {
			jsonRaw:   json.RawMessage(fmt.Sprintf("{\"url\":\"%s\"}", "https://bix24852.dev.dynatracelabs.com")),
			env:       map[string]string{"K6_DYNATRACE_INSECURE_SKIP_TLS_VERIFY": "d"},
			arg:       "",
			config:    Config{},
			errString: "strconv.ParseBool",
		},
		"with_headers_json": {
			jsonRaw: json.RawMessage(fmt.Sprintf("{\"url\":\"%s\", \"headers\":{\"X-Header\":\"value\"}}", "https://bix24852.dev.dynatracelabs.com")),
			env:     nil,
			arg:     "",
			config: Config{
				Url:                   "https://bix24852.dev.dynatracelabs.com",
				InsecureSkipTLSVerify: null.BoolFrom(true),
				CACert:                null.NewString("", false),
				ApiToken:              null.NewString("", false),
				FlushPeriod:           types.NullDurationFrom(defaultFlushPeriod),
				KeepTags:              null.BoolFrom(true),
				KeepNameTag:           null.BoolFrom(false),
				KeepUrlTag:            null.BoolFrom(true),
				Headers: map[string]string{
					"X-Header": "value",
				},
			},
			errString: "",
		},
	}

	for name, testCase := range testCases {
		testCase := testCase
		t.Run(name, func(t *testing.T) {
			c, err := GetConsolidatedConfig(testCase.jsonRaw, testCase.env, testCase.arg)
			if len(testCase.errString) > 0 {
				assert.Contains(t, err.Error(), testCase.errString)
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, testCase.config.Url, c.Url)
			assert.Equal(t, testCase.config.InsecureSkipTLSVerify, c.InsecureSkipTLSVerify)
			assert.Equal(t, testCase.config.CACert, c.CACert)
			assert.Equal(t, testCase.config.ApiToken, c.ApiToken)
			assert.Equal(t, testCase.config.FlushPeriod, c.FlushPeriod)
			assert.Equal(t, testCase.config.KeepTags, c.KeepTags)
			assert.Equal(t, testCase.config.KeepNameTag, c.KeepNameTag)
			assert.Equal(t, testCase.config.KeepUrlTag, c.KeepUrlTag)
			assert.Equal(t, testCase.config.Headers, c.Headers)
		})
	}
}
