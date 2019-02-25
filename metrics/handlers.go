// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"net/http"
	"net/http/pprof"
	"strings"

	"github.com/Stackdriver/heapster/metrics/api/v1"
	metricsApi "github.com/Stackdriver/heapster/metrics/apis/metrics"
	"github.com/Stackdriver/heapster/metrics/core"
	metricsink "github.com/Stackdriver/heapster/metrics/sinks/metric"
	"github.com/Stackdriver/heapster/metrics/util/metrics"
	restful "github.com/emicklei/go-restful"

	v1listers "k8s.io/client-go/listers/core/v1"
)

const pprofBasePath = "/debug/pprof/"

func setupHandlers(metricSink *metricsink.MetricSink, podLister v1listers.PodLister, nodeLister v1listers.NodeLister, historicalSource core.HistoricalSource, disableMetricExport bool) http.Handler {

	runningInKubernetes := true

	// Make API handler.
	wsContainer := restful.NewContainer()
	wsContainer.EnableContentEncoding(true)
	wsContainer.Router(restful.CurlyRouter{})
	a := v1.NewApi(runningInKubernetes, metricSink, historicalSource, disableMetricExport)
	a.Register(wsContainer)
	// Metrics API
	m := metricsApi.NewApi(metricSink, podLister, nodeLister)
	m.Register(wsContainer)

	handlePprofEndpoint := func(req *restful.Request, resp *restful.Response) {
		name := strings.TrimPrefix(req.Request.URL.Path, pprofBasePath)
		switch name {
		case "profile":
			pprof.Profile(resp, req.Request)
		case "symbol":
			pprof.Symbol(resp, req.Request)
		case "cmdline":
			pprof.Cmdline(resp, req.Request)
		default:
			pprof.Index(resp, req.Request)
		}
	}

	// Setup pporf handlers.
	ws := new(restful.WebService).Path(pprofBasePath)
	ws.Route(ws.GET("/{subpath:*}").To(metrics.InstrumentRouteFunc("pprof", handlePprofEndpoint))).Doc("pprof endpoint")
	wsContainer.Add(ws)

	return wsContainer
}
