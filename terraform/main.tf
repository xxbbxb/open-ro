terraform {
  required_version = ">= 1.0"
  required_providers {
    civo = {
      source = "civo/civo"
      version = ">= 1.0.21"
    }
  }
}

variable civo_token {}

provider "civo" {
    token = var.civo_token
    region = "FRA1"
}


resource "civo_firewall" "openro-firewall" {
    name = "openro-firewall"
}

resource "civo_firewall_rule" "openro-kubeapi" {
    firewall_id = civo_firewall.opero-firewall.id
    protocol = "tcp"
    start_port = "6443"
    end_port = "6443"
    cidr = ["0.0.0.0/0"]
    direction = "ingress"
    label = "kubernetes-api-server"
    action = "allow"
}

resource "civo_kubernetes_cluster" "openro" {
    name = "openro"
    firewall_id = civo_firewall.openro-firewall.id
    pools {
        label = "openro-worker"
        size = element(data.civo_size.xsmall.sizes, 0).name
        node_count = 1
    }
}