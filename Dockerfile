# Copyright 2022 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM golang:1.22.4-alpine3.19@sha256:c46c4609d3cc74a149347161fc277e11516f523fd8aa6347c9631527da0b7a56 AS builder

WORKDIR /src
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY ./ ./
RUN go build -o osv-scanner ./cmd/osv-scanner/

FROM alpine:3.20@sha256:b89d9c93e9ed3597455c90a0b88a8bbb5cb7188438f70953fede212a0c4394e0

RUN apk --no-cache add ca-certificates git && \
  git config --global --add safe.directory '*'

WORKDIR /root/
COPY --from=builder /src/osv-scanner .

ENTRYPOINT ["/root/osv-scanner"]
