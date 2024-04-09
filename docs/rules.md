## Transform rules

When generating YAML from XML (or XML from YAML) the following basic rules are used:

* XML elements are represented as YAML fields
* XML elements attributes are represented as YAML fields prepended with a dot `.`
* XML elements content may be represented as a YAML field prepended with `.@` (when needed)
* XML elements that are simple like this `<Simple>Value</Simple>` are represented as `Simple: "Value"`
* XML element sequences are represented as arrays (to preserve order)
* XML elements contents with multiple order-sensitive elements are represented as arrays
* XML elements contents with multiple non-order-sensitive elements are represented as objects

This format is similar to Badgerfish style, but it's not as strict.
It makes concessions so that simple XML translates into simple YAML when possible.

The idea is that the YAML should be intuitive to write by just looking at the XML.

There should not be any loss of information in the transformations (with one exception).
If there is char-data intermingled between XML elements, that is not preserved.

There is no name for this format, you can call it `apigeek-style`


See the examples below

* *Example 1*: XML element containing another XML element
    *  ```xml
       <Parent>
         <Child>foo</Child>
       </Parent>
       ```
       is equivalent to
       ```yaml
       Parent:
         Child: foo
       ```
* *Example 2*: Simple XML element with no attributes, and scalar content
    * ```xml
      <Field>Content</Field>
      ```
      is equivalent to
      ```yaml
      Field: Content
      ```
* *Example 3*: XML element with an attribute
    * ```xml
      <Parent foo="bar" />
      ```
      is equivalent to
      ```yaml
      Parent: 
        .foo: bar
      ```
* *Example 4*: XML element with an attribute and  scalar content
    * ```xml
      <Parent foo="bar" >Content</Parent>
      ``` 
      is equivalent to
      ```yaml
      Parent:
        .foo: bar
        .@: Content
      ```
* *Example 5*: XML sequence where parent has no attributes
    * ```xml
      <Parent>
        <Child>foo</Child>
        <Child>bar</Child>
      </Parent>
      ```
      is equivalent to
      ```yaml
      Parent:
        - Child: foo
        - Child: bar
      ```
* *Example 6*: XML sequence where parent has attributes
    * ```xml
      <Parent attr1="value1" attr2="value2" >
        <Child>foo</Child>
        <Child>bar</Child>
      </Parent>
      ``` 
      is equivalent to
      ```yaml
      Parent:
        .attr1: value1
        .attr2: value2
        .@:
          - Child: foo
          - Child: bar
      ```
* *Example 7*: XML sequence without parent
    * ```xml
      <Root>
        <Child name="foo" />
        <Child name="bar" />
      </Root>
      ```
      is equivalent to
      ```yaml
      Root:
        .@:
          - Child:
            .name: foo
          - Child:
            .name: bar
      ```