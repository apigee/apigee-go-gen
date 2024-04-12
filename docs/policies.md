## Apigee Policies

All Apigee policies are supported, but there is no schema validation on the policies.

You can copy policy XML from the Apigee docs, or from the Apigee UI, and then use 
the `xml2yaml` tool to generate the equivalent YAML.


## Apigee policies sample YAMLs

Below are several examples for common Apigee policies represented as YAML

### Flow Callout

Example Flow Callout policy as YAML.

```yaml
FlowCallout:
  .async: false
  .name: FC-Callout
  .enabled: true
  .continueOnError: true
  DisplayName: FC-Callout
  SharedFlowBundle: SharedFlowName
  Parameters:
    - Parameter:
        .name: param1
        .value: Literal
    - Parameter:
        .name: param2
        .ref: request.content
```

is equivalent to

```text
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<FlowCallout async="false" continueOnError="true" enabled="true" name="FC-Callout" >
  <DisplayName>FC-Callout</DisplayName>
  <Parameters>
    <Parameter name="param1" value="Literal" ></Parameter>
    <Parameter name="param2" ref="request.content" ></Parameter>
  </Parameters>
  <SharedFlowBundle>SharedFlowName</SharedFlowBundle>
</FlowCallout>
```


### Raise Fault

Example Raise Fault policy represented as YAML
```yaml
RaiseFault:
  .async: false
  .name: RF-Example
  .enabled: true
  .continueOnError: true
  DisplayName: RF-Example
  IgnoreUnresolvedVariables: true
  ShortFaultReason: false
  FaultResponse:
    - AssignVariable:
        Name: flow.var
        Value: 123
    - Add:
        Headers:
          - Header:
              .name: user-agent
              -Data: example
    - Copy:
        .source: request
        Headers:
          - Header:
              .name: header-name
        StatusCode: 304
    - Remove:
        Headers:
          - Header:
              .name: sample-header
    - Set:
        Headers:
          - Header:
              .name: user-agent
              -Data: "{request.header.user-agent}"
        Payload:
          .contentType: application/json
          -Data: '{"name":"foo", "type":"bar"}'
    - Set:
        ReasonPhrase: Server Error
        StatusCode: 500

```

is equivalent to

```xml
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<RaiseFault async="false" continueOnError="true" enabled="true" name="RF-Example" >
    <DisplayName>RF-Example</DisplayName>
    <FaultResponse>
        <AssignVariable >
            <Name>flow.var</Name>
            <Value>123</Value>
        </AssignVariable>
        <Add >
            <Headers>
                <Header name="user-agent" >example</Header>
            </Headers>
        </Add>
        <Copy source="request" >
            <Headers>
                <Header name="header-name" ></Header>
            </Headers>
            <StatusCode>304</StatusCode>
        </Copy>
        <Remove >
            <Headers>
                <Header name="sample-header" ></Header>
            </Headers>
        </Remove>
        <Set >
            <Headers>
                <Header name="user-agent" >{request.header.user-agent}</Header>
            </Headers>
            <Payload contentType="application/json" >{"name":"foo", "type":"bar"}</Payload>
        </Set>
        <Set >
            <ReasonPhrase>Server Error</ReasonPhrase>
            <StatusCode>500</StatusCode>
        </Set>
    </FaultResponse>
    <IgnoreUnresolvedVariables>true</IgnoreUnresolvedVariables>
    <ShortFaultReason>false</ShortFaultReason>
</RaiseFault>

```

