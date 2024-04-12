## Transform rules

When generating YAML from XML (or XML from YAML) the following basic rules are used:

* XML elements are represented as YAML fields
* XML attributes are represented as YAML fields prepended with a dot `.`
* XML elements with simple char-data like this `<Simple>Value</Simple>` should be represented as `Simple: Value`
* XML element sequences must be represented as arrays to preserve order
* XML elements containing order-sensitive children must use an array to hold the children
* XML elements containing non-order-sensitive (unique) children should use a map to hold the children
* XML elements having attributes and char-data must put the char-data content within a field prepended with `-`
* XML elements having attributes and order-sensitive children must put children within a field prepended with `-`



This format is similar to Badgerfish style, but it's not as strict.
It makes concessions so that simple XML translates into simple YAML when possible.

The idea is that the YAML should be intuitive to write by just looking at the XML.

There should not be any loss of information in the transformations (with one exception).
If there is char-data intermingled between XML elements, that is not preserved.

There is no name for this format, you can call it `apigeek-style`


See the examples below

* *Example 1*: XML element with char data content
    * ```xml
      <Book>The Cat in the Hat</Book>
      ```
      is equivalent to
      ```yaml
      Book: The Cat in the Hat
      ```
* *Example 2*: XML element with an attribute
    * ```xml
      <Book author="Dr. Seuss" />
      ```
      is equivalent to
      ```yaml
      Book: 
        .author: Dr. Seuss
      ```      
* *Example 3*: XML element with an attribute and char data content
    * ```xml
      <Book author="Dr. Seuss">The Cat in the Hat</Book>
      ``` 
      is equivalent to
      ```yaml
      Book:
        .author: en
        -Data: The Cat in the Hat
      ```      
* *Example 4*: XML element containing another XML element
    *  ```xml
       <Catalog>
         <Book>The Cat in the Hat</Book>
       </Catalog>
       ```
       is equivalent to
       ```yaml
       Catalog:
         Book: The Cat in the Hat
       ```

* *Example 5*: XML sequence that has a container element
    * ```xml
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
* *Example 6*: XML sequence without container element
    * ```xml
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
* *Example 7*: XML sequence without container, but parent has attributes
    * ```xml
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
* *Example 8*: XML sequence without container, but parent has attributes, and children have attributes
    * ```xml
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

* *Example 9*: XML sequence with a container, parent has attributes, and children have attributes
    * ```xml
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
