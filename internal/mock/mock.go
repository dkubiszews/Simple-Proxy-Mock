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
  "log"
  "net/http"
  "sync"
)

// Struct representing mocked endpoints.
type Mock struct {
  mockMap map[string]mockResponse
  rwMockMapMutex sync.RWMutex
}

// Handles mock operations.
func (this *Mock) HandleMockRequest(uriPath string, responseWriter http.ResponseWriter, request *http.Request) bool {
  this.rwMockMapMutex.RLock()
  value, status := this.mockMap[uriPath]
  this.rwMockMapMutex.RUnlock()

  // TODO: use router?
  // TODO: make mockSettings common.
  if status {
    this.rwMockMapMutex.Lock()
    value.handleMock(responseWriter, request)
    this.rwMockMapMutex.Unlock()
  } else if uriPath == "/mockSettings/set" {
    this.handleSetMock(responseWriter, request)
  } else if uriPath == "/mockSettings/clear" {
    this.handleClearMock(responseWriter, request)
  } else if uriPath == "/mockSettings/clearAll" {
    this.handleClearAllMock(responseWriter, request)
  } else if uriPath == "/mockSettings/ping" {

  } else {
    return false
  }
  return true
}

// Creates new mock.
func NewMock() *Mock {
  mockResult := new(Mock)

  mockResult.rwMockMapMutex.Lock()
  mockResult.mockMap = make(map[string]mockResponse)
  mockResult.rwMockMapMutex.Unlock()

  return mockResult
}

func (this *Mock) handlePing(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("pong"))
  w.WriteHeader(http.StatusOK)
}

func (this *Mock) handleSetMock(w http.ResponseWriter, r *http.Request) {
  log.Print("Handle set mock")
  jsonDecoder := json.NewDecoder(r.Body)
  var mockResponse mockResponse
  err := jsonDecoder.Decode(&mockResponse)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  this.rwMockMapMutex.Lock()
  this.mockMap[mockResponse.Endpoint] = mockResponse
  this.rwMockMapMutex.Unlock()

  w.WriteHeader(http.StatusOK)
}

func (this *Mock) handleClearMock(w http.ResponseWriter, r *http.Request) {
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

  this.rwMockMapMutex.Lock()
  delete(this.mockMap, clearMockData.Endpoint)
  this.rwMockMapMutex.Unlock()

  w.WriteHeader(http.StatusOK)
}

func (this *Mock) handleClearAllMock(w http.ResponseWriter, r *http.Request) {
  log.Print("Handle clearAll mock")
  // TODO: add some body to this request

  this.rwMockMapMutex.Lock()
  this.mockMap = make(map[string]mockResponse)
  this.rwMockMapMutex.Unlock()

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
