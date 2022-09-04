terraform {
  required_version = ">= 1.0"
  required_providers {
    civo = {
      source = "civo/civo"
      version = ">= 1.0.21"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 3.22.0"
    }
  }
}

terraform {
  backend "remote" {
    hostname = "xxbbxb.scalr.io"
    organization = "env-u77fmbj0qv4r6qg"
    workspaces {
      name = "open-ro-prod"
    }
  }
}

provider "civo" {
  token = var.civo_token
  region = "FRA1"
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

resource "civo_firewall" "open-ro-firewall" {
    name = "open-ro-firewall"
    create_default_rules = false
}

resource "civo_firewall_rule" "openro-kubeapi" {
    firewall_id = civo_firewall.open-ro-firewall.id
    protocol = "tcp"
    start_port = "6443"
    end_port = "6443"
    cidr = ["0.0.0.0/0"]
    direction = "ingress"
    label = "kubernetes-api-server"
    action = "allow"
}

resource "civo_firewall_rule" "tcp-egress-allow" {
    firewall_id = civo_firewall.open-ro-firewall.id
    protocol = "tcp"
    start_port = "1"
    end_port = "65535"
    cidr = ["0.0.0.0/0"]
    direction = "egress"
    label = "security"
    action = "allow"
}

resource "civo_firewall_rule" "udp-egress-allow" {
    firewall_id = civo_firewall.open-ro-firewall.id
    protocol = "udp"
    start_port = "1"
    end_port = "65535"
    cidr = ["0.0.0.0/0"]
    direction = "egress"
    label = "security"
    action = "allow"
}

resource "civo_firewall_rule" "icmp-ingress-allow" {
    firewall_id = civo_firewall.open-ro-firewall.id
    protocol = "icmp"
    start_port = "1"
    end_port = "65535"
    cidr = ["0.0.0.0/0"]
    direction = "ingress"
    label = "security"
    action = "allow"
}

resource "civo_firewall_rule" "icmp-egress-allow" {
    firewall_id = civo_firewall.open-ro-firewall.id
    protocol = "icmp"
    start_port = "1"
    end_port = "65535"
    cidr = ["0.0.0.0/0"]
    direction = "egress"
    label = "security"
    action = "allow"
}

data "civo_size" "small" {
    filter {
        key = "name"
        values = ["g4s.kube.small"]
        match_by = "re"
    }

    filter {
        key = "type"
        values = ["Kubernetes"]
    }

}

resource "civo_kubernetes_cluster" "open-ro" {
    name = "openro"
    firewall_id = civo_firewall.open-ro-firewall.id
    pools {
        label = "openro-worker"
        size = element(data.civo_size.small.sizes, 0).name
        node_count = 1
    }
}

resource "cloudflare_zone" "open-ro" {
  name = "open-ro.com"
  account_id = var.cloudflare_account_id
}