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


terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
      # This module is designed to accept two AWS providers,
      # which it will refer to as 'aws.primary' and 'aws.secondary'.
      configuration_aliases = [aws.primary, aws.secondary]
    }
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

variable "primary_bucket_name" {
  type = string
}
variable "secondary_bucket_name" {
  type = string
}
variable "gcs_bucket_name" {
  type = string
}


resource "aws_s3_bucket" "primary" {
  provider = aws.primary
  bucket   = var.primary_bucket_name
}

resource "aws_s3_bucket" "secondary" {
  provider = aws.secondary
  bucket   = var.secondary_bucket_name
}

resource "google_storage_bucket" "gcs_bucket" {
  provider = google # 'provider = google' is technically optional for default
  name     = var.gcs_bucket_name
  location = "US-CENTRAL1"
}