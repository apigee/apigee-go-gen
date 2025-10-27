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

string_attr   = "simple string"
number_attr   = 123
bool_attr     = false
null_attr     = null
wrapped_var   = var.foo
wrapped_func  = format("Hello, %s", var.name)
wrapped_cond  = var.a ? "b" : "c"
template_str  = "Hello, ${var.name}!"
list_attr     = [1, "two", true, var.three]
object_attr = {
  naked_key   = "value1"
  "quoted_key" = "value2"
  (var.key)   = "value3"
  nested_list = [var.a, var.b]
}
