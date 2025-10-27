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

data "aws_ami" "example" {
  provider     = aws.west
  most_recent  = true
}

resource "aws_instance" "main" {
  provider     = aws.east
  ami          = data.aws_ami.example.id
  depends_on   = [aws_s3_bucket.foo, "aws_iam_role.bar"]

  lifecycle {
    create_before_destroy = true
    ignore_changes        = [aws_security_group.sg.id, "tags"]
  }

  connection {
    type     = "ssh"
    user     = "admin"
  }

  provisioner "remote-exec" {
    connection {
      type = var.conn_type
    }
    inline = ["echo 'hello'"]
  }
}

resource "aws_instance" "secondary" {
  lifecycle {
    ignore_changes = all
  }
}