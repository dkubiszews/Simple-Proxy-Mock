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
  "encoding/json"
  "net/http"
  "log"
  "os"
)

type MockResponse struct {
  Endpoint string
  StatusCode int
  Body string
}

type RuntimeConfiguration struct {
  mockMap map[string]MockResponse
}

func handleMock(mockResponse MockResponse, w http.ResponseWriter, r *http.Request) {
  log.Print("Handle mocked request")
  w.WriteHeader(mockResponse.StatusCode)
  w.Write([]byte(mockResponse.Body))
}

func handleSetMock(config *RuntimeConfiguration, w http.ResponseWriter, r *http.Request) {
  log.Print("Handle set mock")
  jsonDecoder := json.NewDecoder(r.Body)
  var mockResponse MockResponse
  err := jsonDecoder.Decode(&mockResponse)
  // TODO: response error code instead of panic
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  config.mockMap[mockResponse.Endpoint] = mockResponse
  w.WriteHeader(http.StatusOK)
}

func handleClearMock(config *RuntimeConfiguration, w http.ResponseWriter, r *http.Request) {
  log.Print("Handle clear mock")
  jsonDecoder := json.NewDecoder(r.Body)
  type ClearMockSchema struct {
    Endpoint string
  }
  var clearMockData ClearMockSchema
  err := jsonDecoder.Decode(&clearMockData)
  // TODO: response error code instead of panic
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  delete(config.mockMap, clearMockData.Endpoint)
  w.WriteHeader(http.StatusOK)
}

func handleClearAllMock(config *RuntimeConfiguration, w http.ResponseWriter, r *http.Request) {
  log.Print("Handle clearAll mock")
  // TODO: add some body to this request 
  config.mockMap = make(map[string]MockResponse)
  log.Print(config)
  w.WriteHeader(http.StatusOK)
}

func proxyHandlerIntern(destinationServer string, config *RuntimeConfiguration, w http.ResponseWriter, r *http.Request) {
  log.Print("Handle request")
  if value, status := config.mockMap[r.RequestURI]; status {
    handleMock(value, w, r)
    return
  } else if r.RequestURI == "/mockSettings/set" {
    handleSetMock(config, w, r)
    return
  } else if r.RequestURI == "/mockSettings/clear" {
    handleClearMock(config, w, r)
    return
  } else if r.RequestURI == "/mockSettings/clearAll"{
    handleClearAllMock(config, w, r)
    return
  }
  client := &http.Client{}
  proxyRequest, err := http.NewRequest(r.Method, destinationServer + r.RequestURI, r.Body)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  for name, value := range r.Header {
    proxyRequest.Header.Set(name, value[0])
  }
  proxyResponse, err := client.Do(proxyRequest)
  r.Body.Close()
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

func provideProxyHandler(destinationServer string) func(http.ResponseWriter, *http.Request) {
  // TODO: add sync for config for mutiple threads
  config := RuntimeConfiguration{}
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
