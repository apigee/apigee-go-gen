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

region = "us-east-1"

resource "aws_vpc" "main-vpc" {
  name = "test-vpc"
  tuple = [1, 2, 3]
  mixed = [1, 2, 1 + 2 ]
}

# Single labeled block
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  # Nested unlabeled block
  tags {
    Name = "main-vpc"
  }

  # Another nested block (multiple)
  subnet {
    cidr = "10.0.1.0/24"
  }
  subnet {
    cidr = "10.0.2.0/24"
  }
}