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
package main

import (
  "io"
  "io/ioutil"
  "encoding/json"
  "net/http"
  "log"
  "os"
  "strings"
  "./internal/httpDecorator"
  "./internal/httpLogger"
)

type MockResponse struct {
  Endpoint string
  Header map[string]string
  StatusCode int
  Body string
}

type Mock struct {
  mockMap map[string]MockResponse
}

func handleMock(mockResponse MockResponse, w http.ResponseWriter, r *http.Request) {
  log.Print("Handle mocked request")
  for keyHeader, valueHeader := range mockResponse.Header {
    w.Header().Add(keyHeader, valueHeader)
  }
  w.WriteHeader(mockResponse.StatusCode)
  w.Write([]byte(mockResponse.Body))
}

func handleSetMock(config *Mock, w http.ResponseWriter, r *http.Request) {
  log.Print("Handle set mock")
  jsonDecoder := json.NewDecoder(r.Body)
  var mockResponse MockResponse
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
  config.mockMap = make(map[string]MockResponse)
  log.Print(config)
  w.WriteHeader(http.StatusOK)
}

func handleProxyRequest(destinationServer string, w http.ResponseWriter, r *http.Request) {
  rBody, err := ioutil.ReadAll(r.Body)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  rNewBodyReader := strings.NewReader(string(rBody[:]))

  client := &http.Client{}
  proxyRequest, err := http.NewRequest(r.Method, destinationServer + r.RequestURI, rNewBodyReader)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  for name, value := range r.Header {
    proxyRequest.Header.Set(name, value[0])
  }
  proxyResponse, err := client.Do(proxyRequest)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  for name, value := range proxyResponse.Header {
    w.Header().Set(name, value[0])
  }
  w.WriteHeader(proxyResponse.StatusCode)
  io.Copy(w, proxyResponse.Body)
  proxyResponse.Body.Close()
}

func proxyHandlerIntern(destinationServer string, config *Mock, w http.ResponseWriter, r *http.Request) {
  log.Print("Handle request")
  httpLogger.LogRequest(r)

  wAccessor := httpDecorator.NewResponseWriterAccessor(r.RequestURI, w)
  if value, status := config.mockMap[r.RequestURI]; status {
    handleMock(value, wAccessor, r)
  } else if r.RequestURI == "/mockSettings/set" {
    handleSetMock(config, wAccessor, r)
  } else if r.RequestURI == "/mockSettings/clear" {
    handleClearMock(config, wAccessor, r)
  } else if r.RequestURI == "/mockSettings/clearAll"{
    handleClearAllMock(config, wAccessor, r)
  } else {
    handleProxyRequest(destinationServer, wAccessor, r)
  }

  httpLogger.LogResponse(wAccessor)
}

func provideProxyHandler(destinationServer string) func(http.ResponseWriter, *http.Request) {
  // TODO: add sync for config for mutiple threads
  config := Mock{}
  config.mockMap = make(map[string]MockResponse)
  return func(w http.ResponseWriter, r *http.Request) {
    proxyHandlerIntern(destinationServer, &config, w, r)
  }
}

func main() {
  argsWithoutProg := os.Args[1:]
  if len(argsWithoutProg) != 2 {
    panic("Invalid arguments, provide args in following way: cmd {port_number} {server_url}")
  }
  port := argsWithoutProg[0]
  serverUrl := argsWithoutProg[1]
  log.Print("Start http server")
  http.HandleFunc("/", provideProxyHandler(serverUrl))
  if err := http.ListenAndServe(":" + port, nil); err != nil {
    panic(err)
  }
}
