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


resource "google_compute_router" "router" {
  project = var.gcp_project_id
  name    = "nat-router"
  network = var.vpc_name
  region  = var.gcp_region
}

resource "google_compute_router_nat" "nat" {
  depends_on = [google_compute_router.router]
  name                               = "my-router-nat"
  router                             = google_compute_router.router.name
  region                             = var.gcp_region
  nat_ip_allocate_option             = "AUTO_ONLY"
  source_subnetwork_ip_ranges_to_nat = "ALL_SUBNETWORKS_ALL_IP_RANGES"

  log_config {
    enable = true
    filter = "ERRORS_ONLY"
  }
}

resource "google_compute_address" "keycloak_ip" {
  project      = var.gcp_project_id
  address_type = "INTERNAL"
  region       = var.gcp_region
  subnetwork   = var.vpc_subnet
  name         = "keycloak-idp-instance-ip"
}

resource "google_compute_instance" "keycloak-idp" {
  depends_on = [google_compute_router_nat.nat]
  name = "keycloak-idp-instance"
  machine_type = "e2-medium"
  network_interface {
    network = var.vpc_name
    subnetwork = var.vpc_subnet
    network_ip = google_compute_address.keycloak_ip.address
  }

  tags = ["keycloak-idp-instance", "https-server", "http-server"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-jammy-v20231030"
    }
  }

  metadata_startup_script = templatefile("./startup.sh", {keycloak_admin  = var.keycloak_admin, keycloak_admin_password = var.keycloak_admin_password})
  allow_stopping_for_update = true
}


resource "google_compute_network_endpoint_group" "neg" {
  name         = "keycloak-instance-neg"
  network      = var.vpc_name
  subnetwork   = var.vpc_subnet
  default_port = "8080"
  zone         = var.gcp_zone
  network_endpoint_type = "GCE_VM_IP_PORT"
}

resource "google_compute_network_endpoint" "keycloak-endpoint" {
  depends_on = [
    google_compute_instance.keycloak-idp,
    google_compute_network_endpoint_group.neg,
    google_compute_address.keycloak_ip]
  instance = google_compute_instance.keycloak-idp.name
  network_endpoint_group = google_compute_network_endpoint_group.neg.name
  port = google_compute_network_endpoint_group.neg.default_port
  ip_address = google_compute_address.keycloak_ip.address
}

resource "google_compute_global_address" "keycloak-idp-external-ip" {
  name = "keycloak-idp-external-ip"
}

resource "google_compute_managed_ssl_certificate" "keycloak-idp-cert" {
  name = "keycloak-idp-cert"

  managed {
    domains = ["${google_compute_global_address.keycloak-idp-external-ip.address}.sslip.io", "${google_compute_global_address.keycloak-idp-external-ip.address}.nip.io"]
  }
}

resource "google_compute_health_check" "keycloak-idp-instance-healthcheck" {
  name = "keycloak-idp-instance-healthcheck"
  timeout_sec        = 1
  check_interval_sec = 1

  tcp_health_check {
    port = "8080"
  }
}

resource "google_compute_backend_service" "keycloak-idp-instance-backend-service" {
  depends_on = [
    google_compute_health_check.keycloak-idp-instance-healthcheck,
    google_compute_network_endpoint_group.neg]
  name        = "keycloak-idp-instance-backend-service"
  protocol    = "HTTP"
  timeout_sec = 10

  health_checks = [google_compute_health_check.keycloak-idp-instance-healthcheck.id]
  backend {
    balancing_mode = "RATE"
    max_rate = 100
    group = google_compute_network_endpoint_group.neg.id
  }
}

resource "google_compute_url_map" "keycloak-idp-lb" {
  depends_on = [google_compute_backend_service.keycloak-idp-instance-backend-service]
  name        = "keycloak-idp-lb"
  default_service = google_compute_backend_service.keycloak-idp-instance-backend-service.id
}

resource "google_compute_target_https_proxy" "keycloak-idp-lb-proxy" {
  depends_on = [google_compute_url_map.keycloak-idp-lb]
  name             = "keycloak-idp-lb-proxy"
  url_map          = google_compute_url_map.keycloak-idp-lb.id
  ssl_certificates = [google_compute_managed_ssl_certificate.keycloak-idp-cert.id]
}

resource "google_compute_global_forwarding_rule" "keycloak-idp-instance-fwd-rule" {
  depends_on = [google_compute_global_address.keycloak-idp-external-ip, google_compute_target_https_proxy.keycloak-idp-lb-proxy ]
  name       = "keycloak-idp-instance-fwd-rule"
  target     = google_compute_target_https_proxy.keycloak-idp-lb-proxy.id
  ip_address = google_compute_global_address.keycloak-idp-external-ip.id
  port_range = 443
}


resource "google_compute_firewall" "allow-keycloak-idp-healthcheck" {
  name    = "allow-keycloak-idp-healthcheck"
  network = var.vpc_name

  direction = "INGRESS"
  source_ranges = ["35.191.0.0/16", "130.211.0.0/22"]
  target_tags = ["keycloak-idp-instance"]

  allow {
    protocol = "tcp"
    ports    = ["8080"]
  }
}