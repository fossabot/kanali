// Copyright (c) 2017 Northwestern Mutual.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package app

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	"github.com/northwesternmutual/kanali/cmd/kanali/app/options"
	"github.com/northwesternmutual/kanali/pkg/chain"
	"github.com/northwesternmutual/kanali/pkg/client/clientset/versioned"
	"github.com/northwesternmutual/kanali/pkg/client/informers/externalversions"
	"github.com/northwesternmutual/kanali/pkg/controller/apikey"
	"github.com/northwesternmutual/kanali/pkg/controller/apikeybinding"
	"github.com/northwesternmutual/kanali/pkg/controller/apiproxy"
	"github.com/northwesternmutual/kanali/pkg/controller/mocktarget"
	"github.com/northwesternmutual/kanali/pkg/crds"
	v2CRDs "github.com/northwesternmutual/kanali/pkg/crds/kanali.io/v2"
	"github.com/northwesternmutual/kanali/pkg/log"
	_ "github.com/northwesternmutual/kanali/pkg/metrics"
	"github.com/northwesternmutual/kanali/pkg/middleware"
	"github.com/northwesternmutual/kanali/pkg/run"
	"github.com/northwesternmutual/kanali/pkg/server"
	"github.com/northwesternmutual/kanali/pkg/store/core/v1"
	"github.com/northwesternmutual/kanali/pkg/tracer"
	"github.com/northwesternmutual/kanali/pkg/utils"
)

func Run(sigCtx context.Context) error {

	ctx, cancel := context.WithCancel(sigCtx)
	logger := log.WithContext(nil)

	crdClientset, kanaliClientset, k8sClientset, err := createClientsets()
	if err != nil {
		logger.Fatal(err.Error())
		return err
	}

	kanaliFactory := externalversions.NewSharedInformerFactory(kanaliClientset, 5*time.Minute)
	k8sFactory := informers.NewSharedInformerFactory(k8sClientset, 5*time.Minute)
	v1.SetGlobalInterface(k8sFactory.Core().V1())

	if err := crds.EnsureCRDs(crdClientset.ApiextensionsV1beta1(),
		v2CRDs.ApiProxyCRD,
		v2CRDs.ApiKeyCRD,
		v2CRDs.ApiKeyBindingCRD,
		v2CRDs.MockTargetCRD,
	); err != nil {
		logger.Fatal(err.Error())
		return err
	} else {
		logger.Info("all customresourcedefinitions successfully created")
	}

	decryptionKey, err := utils.LoadDecryptionKey(viper.GetString(options.FlagPluginsAPIKeyDecriptionKeyFile.GetLong()))
	if err != nil {
		logger.Fatal(err.Error())
		return err
	}

	tracer, tracerErr := tracer.New()
	if tracerErr != nil {
		logger.Warn(tracerErr.Error())
	}

	gatewayServer := server.Prepare(&server.Options{
		Name:         "gateway",
		InsecureAddr: viper.GetString(options.FlagServerInsecureBindAddress.GetLong()),
		SecureAddr:   viper.GetString(options.FlagServerSecureBindAddress.GetLong()),
		InsecurePort: viper.GetInt(options.FlagServerInsecurePort.GetLong()),
		SecurePort:   viper.GetInt(options.FlagServerSecurePort.GetLong()),
		TLSKey:       viper.GetString(options.FlagTLSKeyFile.GetLong()),
		TLSCert:      viper.GetString(options.FlagTLSCertFile.GetLong()),
		TLSCa:        viper.GetString(options.FlagTLSCaFile.GetLong()),
		Handler: chain.New().Add(
			middleware.Recorder,
			middleware.Correlation,
			middleware.Metrics,
		).Link(middleware.Gateway),
		Logger: logger.Sugar(),
	})

	profilingServer := server.Prepare(&server.Options{
		Name:         "profiling",
		InsecureAddr: viper.GetString(options.FlagProfilingInsecureBindAddress.GetLong()),
		InsecurePort: viper.GetInt(options.FlagProfilingInsecurePort.GetLong()),
		Handler:      server.ProfilingHandler(),
		Logger:       logger.Sugar(),
	})

	metricsServer := server.Prepare(&server.Options{
		Name:         "prometheus",
		InsecureAddr: viper.GetString(options.FlagPrometheusServerBindAddress.GetLong()),
		InsecurePort: viper.GetInt(options.FlagPrometheusServerPort.GetLong()),
		Handler:      promhttp.Handler(),
		Logger:       logger.Sugar(),
	})

	var g run.Group

	g.Add(ctx, apiproxy.NewApiProxyController(kanaliFactory.Kanali().V2().ApiProxies()))
	g.Add(ctx, apikey.NewApiKeyController(kanaliFactory.Kanali().V2().ApiKeys(), kanaliClientset, decryptionKey))
	g.Add(ctx, apikeybinding.NewApiKeyBindingController(kanaliFactory.Kanali().V2().ApiKeyBindings()))
	g.Add(ctx, mocktarget.NewMockTargetController(kanaliFactory.Kanali().V2().MockTargets()))
	g.Add(ctx, run.InformerWrapper(k8sFactory.Core().V1().Services().Informer()))
	g.Add(ctx, run.InformerWrapper(k8sFactory.Core().V1().Secrets().Informer()))

	if tracerErr == nil {
		g.Add(ctx, tracer)
	}

	g.Add(ctx, metricsServer)
	g.Add(ctx, gatewayServer)

	if viper.GetBool(options.FlagProfilingEnabled.GetLong()) {
		g.Add(ctx, profilingServer)
	}

	g.Add(ctx, run.MonitorContext(cancel))

	return g.Run()
}

func createClientsets() (
	crdClientset *clientset.Clientset,
	kanaliClientset *versioned.Clientset,
	k8sClientset *kubernetes.Clientset,
	err error,
) {
	config, err := utils.GetRestConfig(viper.GetString(options.FlagKubernetesKubeConfig.GetLong()))
	if err != nil {
		return nil, nil, nil, err
	}

	crdClientset, err = clientset.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	kanaliClientset, err = versioned.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	k8sClientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, nil, err
	}

	return crdClientset, kanaliClientset, k8sClientset, nil
}
