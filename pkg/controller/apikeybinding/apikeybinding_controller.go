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

package apikeybinding

import (
	"github.com/northwesternmutual/kanali/pkg/apis/kanali.io/v2"
	"github.com/northwesternmutual/kanali/pkg/log"
	store "github.com/northwesternmutual/kanali/pkg/store/kanali/v2"
	"github.com/northwesternmutual/kanali/pkg/tags"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/cache"
)

type apiKeyBindingController struct{}

func NewController() cache.ResourceEventHandler {
	return &apiKeyBindingController{}
}

func (ctlr *apiKeyBindingController) OnAdd(obj interface{}) {
	logger := log.WithContext(nil)
	binding, ok := obj.(*v2.ApiKeyBinding)
	if !ok {
		logger.Error("received malformed ApiKeyBinding from k8s apiserver")
		return
	}
	store.ApiKeyBindingStore().Set(binding)
	logger.Debug("added ApiKeyBinding",
		zap.String(tags.KanaliApiKeyBindingName, binding.GetName()),
		zap.String(tags.KanaliApiKeyBindingNamespace, binding.GetNamespace()),
	)
}

func (ctlr *apiKeyBindingController) OnUpdate(old interface{}, new interface{}) {
	logger := log.WithContext(nil)
	newBinding, okNew := new.(*v2.ApiKeyBinding)
	oldBinding, okOld := old.(*v2.ApiKeyBinding)
	if !okNew || !okOld {
		logger.Error("received malformed ApiKeyBinding from k8s apiserver")
		return
	}
	store.ApiKeyBindingStore().Update(oldBinding, newBinding)
	logger.With(
		zap.String(tags.KanaliApiKeyBindingName, newBinding.GetName()),
		zap.String(tags.KanaliApiKeyBindingNamespace, newBinding.GetNamespace()),
	).Debug("updated ApiKeyBinding")
}

func (ctlr *apiKeyBindingController) OnDelete(obj interface{}) {
	logger := log.WithContext(nil)
	binding, ok := obj.(*v2.ApiKeyBinding)
	if !ok {
		logger.Error("received malformed ApiKeyBinding from k8s apiserver")
		return
	}
	if err := store.ApiKeyBindingStore().Delete(binding); err == nil {
		logger.With(
			zap.String(tags.KanaliApiKeyBindingName, binding.GetName()),
			zap.String(tags.KanaliApiKeyBindingNamespace, binding.GetNamespace()),
		).Debug("deleted ApiKeyBinding")
	}
}
