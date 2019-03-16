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
package httpLogger

import (
  "compress/gzip"
  "fmt"
  "io/ioutil"
  "log"
  "net/http"
  "strings"
  "../httpDecorator"
)

// Log http.Request to terminal
// Decompressing gziped body.
func LogRequest(r *http.Request) {
  var sb strings.Builder
  sb.WriteString("\n>>>> REQUEST: " + r.RequestURI + "\n\nHEADER:\n")
  for keyHeader, valueHeader := range r.Header {
    sb.WriteString(keyHeader + ": "  + valueHeader[0] + "\n")
  }

  rBody, _ := ioutil.ReadAll(r.Body)
  sb.WriteString("\nBODY:\n")
  if r.Header.Get("Content-Encoding") == "gzip" {
    // TODO: should react on errors here?
    rNewReader := strings.NewReader(string(rBody[:]))
    gzipReader, _ := gzip.NewReader(rNewReader)
    rUncompressedBody, _ := ioutil.ReadAll(gzipReader)
    sb.WriteString("(uncompressed body)\n" + string(rUncompressedBody[:]))
  } else {
    sb.WriteString(string(rBody[:]))
  }
  sb.WriteString("\n")
  r.Body = ioutil.NopCloser(strings.NewReader(string(rBody[:])))

  log.Print(sb.String())
}

// TODO: consider make common with logRequest.
// Log httpDecorator.ResponseWriterAccessor to terminal.
// Decompressing gziped body.
func LogResponse(r *httpDecorator.ResponseWriterAccessor) {
  var sb strings.Builder
  sb.WriteString("\n<<<< RESPONSE: " + r.RequestURI + "\n\nHEADER:\n")
  for keyHeader, valueHeader := range r.Header() {
    sb.WriteString(keyHeader + ": "  + valueHeader[0] + "\n")
  }

  sb.WriteString("\nBODY:\n")
  if r.Header().Get("Content-Encoding") == "gzip" {
    // TODO: should react on errors here?
    rNewReader := strings.NewReader(r.Body)
    gzipReader, _ := gzip.NewReader(rNewReader)
    rUncompressedBody, _ := ioutil.ReadAll(gzipReader)
    sb.WriteString("(uncompressed body)\n" + string(rUncompressedBody[:]))
  } else {
    sb.WriteString(r.Body)
  }

  sb.WriteString(fmt.Sprintf("STATUS_CODE: %d \n", r.StatusCode))
  sb.WriteString("\n")
  log.Print(sb.String())
}