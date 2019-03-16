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
package httpDecorator

import (
	"net/http"
)

// Decorate http.ResponseWriter to have access to attributes.
type ResponseWriterAccessor struct {
	// Decorated http.ResponseWriter.
	respWriter http.ResponseWriter

	// URI to related request.
	RequestURI string
	
	// Body of the response.
	Body string

	// Status code for the response.
	StatusCode int
}

// decorated method from http.ResponseWriter interface.
func (this *ResponseWriterAccessor) Header() http.Header {
	return this.respWriter.Header()
}

// decorated method from http.ResponseWriter interface.
func (this *ResponseWriterAccessor) Write(data []byte) (int, error) {
	this.Body = string(data)
	return this.respWriter.Write(data)
}

// decorated method from http.ResponseWriter interface.
func (this *ResponseWriterAccessor) WriteHeader(statusCode int) {
	this.StatusCode = statusCode
	this.respWriter.WriteHeader(statusCode)
}

// Creates new decorated response writter for given RequestURI and http.ResponseWriter.
func NewResponseWriterAccessor(requestURI string, respWriter http.ResponseWriter) (*ResponseWriterAccessor) {
  object := new(ResponseWriterAccessor)
  object.RequestURI = requestURI
	object.respWriter = respWriter
	object.StatusCode = http.StatusOK
	return object
}