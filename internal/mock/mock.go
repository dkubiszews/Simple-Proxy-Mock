////////////////////////////////////////////////////////////////////////////
//
// Copyright 2019 Dawid Kubiszewski (dawidkubiszewski@gmail.com)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
////////////////////////////////////////////////////////////////////////////
package mock

import (
	"encoding/json"
	"net/http"
  "log"
)

type Mock struct {
  mockMap map[string]mockResponse
}

func (this *Mock) HandleMockRequest(uriPath string, responseWriter http.ResponseWriter, request *http.Request) bool {
  if value, status := this.mockMap[uriPath]; status {
    value.handleMock(responseWriter, request)
  } else if uriPath == "/mockSettings/set" {
    handleSetMock(this, responseWriter, request)
  } else if uriPath == "/mockSettings/clear" {
    handleClearMock(this, responseWriter, request)
  } else if uriPath == "/mockSettings/clearAll"{
    handleClearAllMock(this, responseWriter, request)
  } else {
    return false;
  }
  return true;
}

func NewMock() (*Mock) {
	mockResult := new(Mock)
	mockResult.mockMap = make(map[string]mockResponse)
	return mockResult
}

func handleSetMock(config *Mock, w http.ResponseWriter, r *http.Request) {
  log.Print("Handle set mock")
  jsonDecoder := json.NewDecoder(r.Body)
  var mockResponse mockResponse
  err := jsonDecoder.Decode(&mockResponse)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  config.mockMap[mockResponse.Endpoint] = mockResponse
  w.WriteHeader(http.StatusOK)
}

func handleClearMock(config *Mock, w http.ResponseWriter, r *http.Request) {
  log.Print("Handle clear mock")
  jsonDecoder := json.NewDecoder(r.Body)
  type ClearMockSchema struct {
    Endpoint string
  }
  var clearMockData ClearMockSchema
  err := jsonDecoder.Decode(&clearMockData)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  delete(config.mockMap, clearMockData.Endpoint)
  w.WriteHeader(http.StatusOK)
}

func handleClearAllMock(config *Mock, w http.ResponseWriter, r *http.Request) {
  log.Print("Handle clearAll mock")
  // TODO: add some body to this request 
  config.mockMap = make(map[string]mockResponse)
  log.Print(config)
  w.WriteHeader(http.StatusOK)
}

type mockResponse struct {
  Endpoint string
  Header map[string]string
  StatusCode int
  Body string
}

func (this *mockResponse) handleMock(w http.ResponseWriter, r *http.Request) {
  log.Print("Handle mocked request")
  for keyHeader, valueHeader := range this.Header {
    w.Header().Add(keyHeader, valueHeader)
  }
  w.WriteHeader(this.StatusCode)
  w.Write([]byte(this.Body))
}