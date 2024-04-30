# Using gRPC Proto
<!--
  Copyright 2024 Google LLC

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.
-->

You can use the [render apiproxy](./commands/render-apiproxy.md) command to create an Apigee API proxy bundle using a template and a gRPC proto file as input.

When working with gRPC in Apigee, it's crucial to ensure your API proxy's base path and conditional flows are configured
correctly to handle gRPC traffic. This command simplifies the process by letting you build a template that understands
these gRPC-specific requirements.


## How it works

- [x] **Start with Your Template**
> Input your gRPC proto file, and the template generates the intermediate YAML configuration.
- [x] **Automate the Details**
> The template handles the intricacies of gRPC integration within your API proxy.
- [x] **Access the Proto**
> Use `--set-grpc` to access the gRPC proto text and [descriptor](https://pkg.go.dev/google.golang.org/protobuf/types/descriptorpb#FileDescriptorProto) during template rendering.
- [x] **Inject Your Values** 
> Use `--set` and `--set-string` to provide missing values (like target server) for your template.


## Example

Check out the [examples/templates/grpc](https://github.com/apigee/apigee-go-gen/blob/main/examples/templates/grpc/apiproxy.yaml) directory for an example of building the intermediate YAML for a gRPC API proxy.

Here is how you would use the [render apiproxy](./commands/render-apiproxy.md) command with this example:

#### Create bundle zip
```shell
apigee-go-gen render apiproxy \
    --template ./examples/templates/grpc/apiproxy.yaml \
    --set-grpc proto=./examples/protos/greeter.proto \
    --set-string "target_server=example-target-server" \
    --include ./examples/templates/grpc/*.tmpl \
    --output ./out/apiproxies/greeter.zip
```

#### Create bundle dir
```shell
apigee-go-gen render apiproxy \
    --template ./examples/templates/grpc/apiproxy.yaml \
    --set-grpc proto=./examples/protos/greeter.proto \
    --set-string "target_server=example-target-server" \
    --include ./examples/templates/grpc/*.tmpl \
    --output ./out/apiproxies/greeter
```