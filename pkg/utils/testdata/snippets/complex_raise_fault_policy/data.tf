#  Copyright 2025 Google LLC
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#       http:#www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

RaiseFault {
  _async = false
  _continueOnError = true
  _enabled = true
  _name = "RF-Example"
  DisplayName = "RF-Example"
  IgnoreUnresolvedVariables = true
  ShortFaultReason = false

  FaultResponse "AssignVariable" {
    Name = "flow.var"
    Value = 123
  }

  FaultResponse "Add" "Headers" "Header" {
    _Data = "example"
    _name = "user-agent"
  }

  FaultResponse "Copy" {
    _source = "request"
    StatusCode = 304

    Headers "Header" {
      _name = "header-name"
    }
  }

  FaultResponse "Remove" "Headers" "Header" {
    _name = "sample-header"
  }

  FaultResponse "Set" {
    Headers "Header" {
      _Data = "{request.header.user-agent}"
      _name = "user-agent"
    }

    Payload {
      _Data = "{\"name\":\"foo\", \"type\":\"bar\"}"
      _contentType = "application/json"
    }
  }

  FaultResponse "Set" {
    ReasonPhrase = "Server Error"
    StatusCode = 500
  }
}