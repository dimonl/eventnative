package events

import (
	"errors"
	"github.com/ksensehq/eventnative/appconfig"
	"github.com/ksensehq/eventnative/geo"
	"github.com/ksensehq/eventnative/logging"
	"github.com/ksensehq/eventnative/useragent"
	"net/http"
)

//ApiPreprocessor preprocess server 2 server integration events
type ApiPreprocessor struct {
	geoResolver geo.Resolver
	uaResolver  useragent.Resolver
}

func NewApiPreprocessor() Preprocessor {
	return &ApiPreprocessor{
		geoResolver: appconfig.Instance.GeoResolver,
		uaResolver:  appconfig.Instance.UaResolver,
	}
}

//Preprocess resolve geo from ip field or skip if geo.GeoDataKey field was provided
//resolve useragent from uaKey or skip if useragent.ParsedUaKey field was provided
//return same object
func (ap *ApiPreprocessor) Preprocess(fact Fact, r *http.Request) (Fact, error) {
	if fact == nil {
		return nil, errors.New("Input fact can't be nil")
	}

	fact["src"] = "api"
	ip := extractIp(r)
	if ip != "" {
		fact[ipKey] = ip
	}

	if deviceCtx, ok := fact["device_ctx"]; ok {
		if deviceCtxObject, ok := deviceCtx.(map[string]interface{}); ok {
			//geo.GeoDataKey node overwrite geo resolving
			if _, ok := deviceCtxObject[geo.GeoDataKey]; !ok {
				if ip, ok := deviceCtxObject["ip"]; ok {
					geoData, err := ap.geoResolver.Resolve(ip.(string))
					if err != nil {
						logging.Error(err)
					}

					deviceCtxObject[geo.GeoDataKey] = geoData
				}
			}

			//useragent.ParsedUaKey node overwrite useragent resolving
			if _, ok := deviceCtxObject[useragent.ParsedUaKey]; !ok {
				if ua, ok := deviceCtxObject[uaKey]; ok {
					if uaStr, ok := ua.(string); ok {
						deviceCtxObject[useragent.ParsedUaKey] = ap.uaResolver.Resolve(uaStr)
					}
				}
			}
		}
	}

	return fact, nil
}
