# XML to YAML
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

This command takes an XML snippet and converts it into YAML.


Let's say you have an Apigee policy written in XML format. Instead of manually retyping the whole thing into YAML, you can simply use this tool for instant conversion. 

This is handy when you're working with examples from the Apigee documentation, just:

1. [x] Copy
2. [x] Paste
3. [x] Convert



## Usage

The `xml-to-yaml` command takes two parameters `-input` and `-output`

* `--input` is the XML document to transform

* `--output` is the YAML document to be created

* `--output` full path is created if it does not exist (like `mkdir -p`)

> You may omit the `--input` or `--output` flags to read or write from stdin or stdout


### Examples
Below are a few examples for using the `xml-to-yaml` command.

#### From a file 
Reading input redirected from a file
```shell
apigee-go-gen transform xml-to-yaml < ./examples/snippets/check-quota.xml
```

#### From stdin
Reading input directly from stdin
```shell
apigee-go-gen transform xml-to-yaml << EOF 
<Parent>
  <Child>Fizz</Child>
  <Child>Buzz</Child>
</Parent>
EOF
```

#### From a process
Reading input piped from another process
```shell
echo '
<Parent>
  <Child>Fizz</Child>
  <Child>Buzz</Child>
</Parent>' | apigee-go-gen transform xml-to-yaml
```


## Transform Rules

When converting between XML and YAML the following basic rules are used:


| XML Representation                                     | YAML Representation                                  |
|--------------------------------------------------------|------------------------------------------------------|
| Element                                                | As Field                                             |
| Attribute                                              | As Field prepended with a dot `.`                    |
| Element with char-data e.g. `<Simple>Value</Simple>`   | As Field with Scalar content e.g. `Simple: Value`    |
| Element sequence                                       | Must use Array to hold the children                  |
| Element with order-sensitive children                  | Must use Array to hold the children                  |
| Element with non-order-sensitive children (unique)     | Should use map to hold the children                  | 
| Element having attributes and char-data                | Must put char-data within a field prepended with `-` | 
| Element having attributes and order-sensitive children | Must put children within a field prepended with `-`  |


This format is similar to [Badgerfish style](http://www.sklar.com/badgerfish/), but it's not as strict.

It makes concessions so that simple XML translates into simple YAML when possible.

The idea is that the YAML should be intuitive to write by just looking at the XML.

There is no name for this format, you can call it `apigeek-style`

!!! Note
    If there is char-data intermingled between XML elements, that is not preserved during transform.


### XML to YAML Examples

Below is a list of examples to help illustrate the XML to YAML transformation logic.

The examples start very simple, and get gradually more complex.

#### 1. Simple element
XML element with char data content
```xml
<Book>The Cat in the Hat</Book>
```
is equivalent to
```yaml
Book: The Cat in the Hat
```

#### 2. Add Attributes
XML element with an attribute
```xml
<Book author="Dr. Seuss" />
```
is equivalent to
```yaml
Book: 
  .author: Dr. Seuss
```     

#### 3. Content & Attributes
XML element with an attribute and char data content
```xml
<Book author="Dr. Seuss">The Cat in the Hat</Book>
```
is equivalent to
```yaml
Book:
  .author: Dr. Seuss
  -Data: The Cat in the Hat
```

#### 4. Nested Elements
XML element containing another XML element
```xml
<Catalog>
 <Book>The Cat in the Hat</Book>
</Catalog>
```
is equivalent to
```yaml
Catalog:
  Book: The Cat in the Hat
```

#### 5. Named Sequence
XML sequence that has a container element
```xml
<Catalog>
  <Books>
    <Book>The Cat in the Hat</Book>
    <Book>Green Eggs and Ham</Book>
  </Books>
</Catalog>
```
is equivalent to
```yaml
Catalog:
  Books:
    - Book: The Cat in the Hat
    - Book: Green Eggs and Ham
```

#### 6. Unnamed Sequence
XML sequence without container element
```xml
<Catalog>
  <Book>The Cat in the Hat</Book>
  <Book>Green Eggs and Ham</Book>
</Catalog>
```
is equivalent to
```yaml
Catalog:
  - Book: The Cat in the Hat
  - Book: Green Eggs and Ham
```

#### 7. Sequence & Attrs.
XML sequence without container, but parent has attributes
```xml
<Catalog name="Children's Books" language="English">
  <Book>The Cat in the Hat</Book>
  <Book>Green Eggs and Ham</Book>
</Catalog>
``` 
is equivalent to
```yaml
Catalog:
  .name: Children's Books
  .language: English
  -Data:
    - Book: The Cat in the Hat
    - Book: Green Eggs and Ham
```

#### 8. Unnamed Seq. & Attrs.
XML sequence without container, but parent has attributes, and children have attributes
```xml
<Catalog name="Children's Books" language="English">
  <Book author="Dr. Seuss">The Cat in the Hat</Book>
  <Book author="Dr. Seuss">Green Eggs and Ham</Book>
</Catalog>
``` 
is equivalent to
```yaml
Catalog:
  .name: Children's Books
  .language: English
  -Data:
    - Book: 
        .author: Dr. Seuss
        -Data: The Cat in the Hat
    - Book:
        .author: Dr. Seuss
        -Data: Green Eggs and Ham
```

#### 9. Named Seq. & Attrs.
XML sequence with a container, parent has attributes, and children have attributes
```xml
<Catalog name="Children's Books" language="English">
  <Books>
    <Book author="Dr. Seuss">The Cat in the Hat</Book>
    <Book author="Dr. Seuss">Green Eggs and Ham</Book>
  </Books>
</Catalog>
``` 
is equivalent to
```yaml
Parent:
  .name: Children's Books
  .language: English
  Books:
    - Book: 
        .author: Dr. Seuss
        -Data: The Cat in the Hat
    - Book:
        .author: Dr. Seuss
        -Data: Green Eggs and Ham
```      
