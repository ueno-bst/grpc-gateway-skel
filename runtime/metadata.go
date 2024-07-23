package runtime

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/metadata"
	"log"
	"net/http"
	"regexp"
	"strings"
)

type MetaType int

const (
	RequestMeta MetaType = iota
	ResponseMeta
	BidirectionalMeta
)

type MetaAction int

const (
	MetaPathThrowAction MetaAction = iota
	MetaAppendAction
	MetaDeleteAction
)

type mdValue struct {
	key     string
	pattern *regexp.Regexp
	value   *string
	action  MetaAction
}

type mdValues []mdValue

func (d mdValues) match(values map[string][]string, prefix string) metadata.MD {
	md := make(metadata.MD)

	for k, v := range values {
		if prefix != "" && strings.HasPrefix(k, prefix) {
			k, _ = strings.CutPrefix(k, prefix)
		}

		for _, i := range d {
			if i.pattern != nil && i.pattern.MatchString(k) {
				switch i.action {
				case MetaPathThrowAction:
					md.Set(k, v...)
					break
				default:
					break
				}

				break
			}
		}
	}

	return md
}

type metadataInfo struct {
	req mdValues
	res mdValues
}

func (i *metadataInfo) add(v mdValue, m MetaType) {
	if m == RequestMeta || m == BidirectionalMeta {
		i.req = append(i.req, v)
	}

	if m == ResponseMeta || m == BidirectionalMeta {
		i.res = append(i.res, v)
	}
}

type WithMetaDataFunc = func(meta metadataInfo) metadataInfo

func WithMetadata(options ...WithMetaDataFunc) GatewayOptionFunc {
	return func(opt *GatewayOption) {
		info := metadataInfo{
			res: mdValues{},
			req: mdValues{},
		}

		for _, option := range options {
			info = option(info)
		}

		opt.metas = info
		opt.muxOpts = append(opt.muxOpts, metadataMuxFunc(opt))

		log.Printf("%+o\n", info)
	}
}

func PassThrowMeta(keys []string, mode MetaType) WithMetaDataFunc {
	return func(info metadataInfo) metadataInfo {
		for _, key := range keys {
			info.add(mdValue{
				key:     key,
				pattern: regexCompile(key),
				action:  MetaPathThrowAction,
			}, mode)
		}

		return info
	}
}

func DeleteMeta(keys []string, mode MetaType) WithMetaDataFunc {
	return func(info metadataInfo) metadataInfo {
		for _, key := range keys {
			info.add(mdValue{
				key:     key,
				pattern: regexCompile(key),
				action:  MetaDeleteAction,
			}, mode)
		}

		return info
	}
}

func metadataMuxFunc(opt *GatewayOption) runtime.ServeMuxOption {
	return runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
		if len(opt.metas.req) > 0 {
			return opt.metas.req.match(req.Header, "")
		}

		return make(metadata.MD)
	})
}

func regexCompile(key string) *regexp.Regexp {
	p := regexp.QuoteMeta(key)

	if strings.Contains(key, "*") {
		p = strings.ReplaceAll(p, "\\*", ".*")
	}

	return regexp.MustCompile(p)
}
